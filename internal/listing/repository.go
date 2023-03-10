package listing

import (
	"context"
	"encoding/base64"
	"fmt"
	thumbnailspb "fuu/v/gen/go/grpc/thumbnails/v1"
	"fuu/v/internal/domain"
	"fuu/v/pkg/instrumentation"
	"time"

	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	rdb    *redis.Client
	logger *zap.SugaredLogger
	conn   *grpc.ClientConn
}

func (r *Repository) Count(ctx context.Context) (int64, error) {
	_, span := otel.Tracer(otelName).Start(ctx, "listing.Count")
	defer span.End()

	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Directory{}).Count(&count).Error
	return count, err
}

func (r *Repository) Create(ctx context.Context, path, name, thumbnail string) (domain.Directory, error) {
	_, span := otel.Tracer(otelName).Start(ctx, "listing.Create")
	defer span.End()

	m := domain.Directory{
		Name:      name,
		Path:      path,
		Thumbnail: thumbnail,
		Loved:     false,
	}
	err := r.db.WithContext(ctx).Create(&m).Error
	return m, err
}

func (r *Repository) FindByName(ctx context.Context, name string) (domain.Directory, error) {
	_, span := otel.Tracer(otelName).Start(ctx, "listing.FindByName")
	defer span.End()

	m := domain.Directory{}
	err := r.db.WithContext(ctx).First(&m, name).Error
	return m, err
}

func (r *Repository) FindAllByName(ctx context.Context, filter string) (*[]domain.Directory, error) {
	_, span := otel.Tracer(otelName).Start(ctx, "listing.FindAllByName")
	defer span.End()

	r.logger.Infow("FindAllByName", "filter", filter)
	all := new([]domain.Directory)

	cached, _ := r.rdb.Get(ctx, filter).Bytes()

	if len(cached) > 0 {
		json.Unmarshal(cached, all)
		instrumentation.CacheHitCounter.Add(1)
		return all, nil
	}

	err := r.db.WithContext(ctx).
		Table("directories").
		Select("id", "name", "loved", "directories.path", "name", "created_at", "updated_at", "thumbnails.thumbnail").
		Joins("left join thumbnails on directories.path = thumbnails.folder").
		Where("name LIKE ?", "%"+filter+"%").
		Find(all).Error

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	encoded, err := json.Marshal(*all)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	redisErr := r.rdb.SetNX(ctx, filter, encoded, time.Minute).Err()
	if err != nil {
		span.RecordError(redisErr)
	}
	instrumentation.CacheMissCounter.Add(1)

	return all, nil
}

func (r *Repository) FindAllRange(ctx context.Context, take, skip, order int) (*[]domain.Directory, error) {
	_, span := otel.Tracer(otelName).Start(ctx, "listing.FindAllRange")
	defer span.End()

	cacheKey := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("findallrange%d%d%d", take, skip, order)),
	)

	r.logger.Infow("FindAllRange", "take", take, "skip", skip)
	_range := new([]domain.Directory)

	cached, _ := r.rdb.Get(ctx, cacheKey).Bytes()

	if len(cached) > 0 {
		json.Unmarshal(cached, _range)
		instrumentation.CacheHitCounter.Add(1)
		return _range, nil
	}

	client := thumbnailspb.NewThumbnailServiceClient(r.conn)

	var _order string
	if order == domain.OrderByDate {
		_order = "updated_at desc"
	}
	if order == domain.OrderByName {
		_order = "name"
	}

	err := r.db.WithContext(ctx).
		Table("directories").
		Select("id", "name", "loved", "directories.path", "name", "created_at", "updated_at", "thumbnails.thumbnail").
		Joins("left join thumbnails on directories.path = thumbnails.folder").
		Order(_order).
		Limit(take).
		Offset(skip).
		Where("thumbnails.thumbnail <> ''").
		Find(_range).Error

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	res, err := client.GetRange(ctx, &thumbnailspb.GetRangeRequest{
		Paths: []string{},
	})

	for _, t := range res.Thumbnails {
		r.logger.Infoln(t.Path, t.Id)
	}

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	encoded, err := json.Marshal(*_range)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	redisErr := r.rdb.SetNX(ctx, cacheKey, encoded, time.Minute).Err()
	if err != nil {
		span.RecordError(redisErr)
	}

	instrumentation.CacheMissCounter.Add(1)

	return _range, err
}

func (r *Repository) FindLikeNameRange(ctx context.Context, filter string, take, skip int) (*[]domain.Directory, int64, error) {
	_, span := otel.Tracer(otelName).Start(ctx, "listing.FindLikeNameRange")
	defer span.End()

	cacheKey := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s%d%d", filter, take, skip)),
	)

	r.logger.Infow("FindLikeNameRange", "filter", filter, "take", take, "skip", skip)
	_range := new([]domain.Directory)

	cached, _ := r.rdb.Get(ctx, cacheKey).Bytes()

	if len(cached) > 0 {
		json.Unmarshal(cached, _range)
		instrumentation.CacheHitCounter.Add(1)
		return _range, 0, nil
	}

	client := thumbnailspb.NewThumbnailServiceClient(r.conn)

	var count int64

	err := r.db.WithContext(ctx).
		Model(&domain.Directory{}).
		Where("name LIKE ?", "%"+filter+"%").
		Count(&count).Error

	if err != nil {
		span.RecordError(err)
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Table("directories").
		Select("id", "name", "loved", "directories.path", "name", "created_at", "updated_at", "thumbnails.thumbnail").
		Joins("left join thumbnails on directories.path = thumbnails.folder").
		Limit(take).
		Offset(skip).
		Where("thumbnails.thumbnail <> '' AND name LIKE ?", "%"+filter+"%").
		Find(_range).Error

	if err != nil {
		span.RecordError(err)
		return nil, 0, err
	}

	res, err := client.GetRange(ctx, &thumbnailspb.GetRangeRequest{
		Paths: []string{},
	})

	for _, t := range res.Thumbnails {
		r.logger.Infoln(t.Path, t.Id)
	}

	if err != nil {
		span.RecordError(err)
		return nil, 0, err
	}

	encoded, err := json.Marshal(*_range)
	if err != nil {
		span.RecordError(err)
		return nil, 0, err
	}

	redisErr := r.rdb.SetNX(ctx, cacheKey, encoded, time.Minute).Err()
	if err != nil {
		span.RecordError(redisErr)
	}

	instrumentation.CacheMissCounter.Add(1)

	return _range, count, err
}

func (r *Repository) FindAll(ctx context.Context) (*[]domain.Directory, error) {
	_, span := otel.Tracer(otelName).Start(ctx, "listing.FindAll")
	defer span.End()

	all := new([]domain.Directory)
	err := r.db.WithContext(ctx).Find(all).Error
	return all, err
}

func (r *Repository) Update(ctx context.Context, path, name, thumbnail string) (domain.Directory, error) {
	_, span := otel.Tracer(otelName).Start(ctx, "listing.Update")
	defer span.End()

	m := domain.Directory{}
	err := r.db.WithContext(ctx).First(&m).Error
	if err != nil {
		span.RecordError(err)
		return domain.Directory{}, err
	}

	m.Name = name
	m.Path = path
	m.Thumbnail = thumbnail
	err = r.db.WithContext(ctx).Save(&m).Error

	return m, err
}

func (r *Repository) Delete(ctx context.Context, path string) (domain.Directory, error) {
	_, span := otel.Tracer(otelName).Start(ctx, "listing.Delete")
	defer span.End()

	m := domain.Directory{}
	err := r.db.WithContext(ctx).Where("path = ?", fmt.Sprintf("`%s`", path)).Delete(&domain.Directory{}).Error
	return m, err
}
