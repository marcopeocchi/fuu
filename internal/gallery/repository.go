package gallery

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"fuu/v/internal/domain"
	"fuu/v/pkg/instrumentation"
	"fuu/v/pkg/utils"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/goccy/go-json"
	"github.com/marcopeocchi/fazzoletti/slices"
	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

const otelName = "fuu/v/internal/gallery"

type Repository struct {
	rdb        *redis.Client
	ch         *amqp.Channel
	logger     *zap.SugaredLogger
	workingDir string
}

func (r *Repository) FindByPath(ctx context.Context, path string) (domain.Content, error) {
	_, span := otel.Tracer(otelName).Start(ctx, "gallery.FindByPath")
	defer span.End()

	cached, _ := r.rdb.Get(ctx, path).Bytes()

	if len(cached) > 0 {
		r.logger.Infow("retrieved cached", "path", path)

		res := domain.Content{}
		err := json.Unmarshal(cached, &res)
		if err != nil {
			span.RecordError(err)
		}
		res.Cached = true

		instrumentation.CacheHitCounter.Add(1)

		return res, err
	}

	start := time.Now()
	r.logger.Infow("accessing filesystem", "path", path)

	wd := filepath.Join(r.workingDir, path)

	files, _ := os.ReadDir(wd)
	filesAvif, _ := os.ReadDir(filepath.Join(wd, "avif"))
	filesWebp, _ := os.ReadDir(filepath.Join(wd, "webp"))

	filterFunc := func(file fs.DirEntry) bool {
		mimeType := mime.TypeByExtension(filepath.Ext(file.Name()))
		return utils.ValidType.MatchString(mimeType) && utils.ValidFile(file.Name())
	}

	files = slices.Filter(files, func(file fs.DirEntry) bool {
		return filterFunc(file)
	})

	filesAvif = slices.Filter(filesAvif, func(file fs.DirEntry) bool {
		return filterFunc(file)
	})

	r.logger.Infow(
		"retrieved resources from filesystem",
		"elapsed", time.Since(start),
	)

	resOrig := make([]string, len(files))
	resAvif := make([]string, len(filesAvif))
	resWebp := make([]string, len(filesWebp))

	for i, file := range files {
		if !file.IsDir() {
			resOrig[i] = file.Name()
		}
	}

	onlyImgs := slices.Filter(resOrig, func(f string) bool {
		return utils.IsImagePath(f)
	})

	for i, file := range filesAvif {
		if !file.IsDir() {
			resAvif[i] = fmt.Sprintf("/avif/%s", file.Name())
		}
	}

	for i, file := range filesWebp {
		if !file.IsDir() {
			resWebp[i] = fmt.Sprintf("/webp/%s", file.Name())
		}
	}

	sort.SliceStable(resOrig, func(i, j int) bool {
		return utils.FilesSortFunc(i, j, resOrig)
	})

	sort.SliceStable(resAvif, func(i, j int) bool {
		return utils.FilesSortFunc(i, j, resAvif)
	})

	sort.SliceStable(resWebp, func(i, j int) bool {
		return utils.FilesSortFunc(i, j, resWebp)
	})

	content := domain.Content{
		Source:        resOrig,
		Avif:          resAvif,
		WebP:          resWebp,
		AvifAvailable: len(resAvif) >= len(onlyImgs),
		WebPAvailable: len(resWebp) >= len(onlyImgs),
	}

	encoded, err := json.Marshal(content)

	if err != nil {
		r.logger.Errorw("encoding error", "error", err)
		span.RecordError(err)
		return domain.Content{}, err
	}

	// Write-through caching
	r.logger.Infow(
		"caching resources",
		"mode", "write-through",
		"ttl", time.Second*30,
		"path", path,
	)
	r.rdb.SetNX(ctx, path, encoded, time.Second*30)
	instrumentation.CacheMissCounter.Add(1)

	// Send images to RabbitMQ for processing
	if len(resWebp) < len(resOrig) {
		// reusable buffer
		var b bytes.Buffer

		for _, image := range resOrig {
			toSend := filepath.Join(wd, image)

			if err := gob.NewEncoder(&b).Encode(toSend); err != nil {
				return domain.Content{}, err
			}

			err := r.ch.Publish(
				"images",                // exchange
				"gallery.event.convert", // routing key
				false,                   // mandatory
				false,                   // immediate
				amqp.Publishing{
					AppId:       "fuu",
					ContentType: "application/x-encoding-gob",
					Body:        b.Bytes(),
					Timestamp:   time.Now(),
				},
			)
			if err != nil {
				span.RecordError(err)
				return domain.Content{}, err
			}
			r.logger.Infow("published message", "msg", image)
			b.Reset()
		}
	}

	return content, nil
}
