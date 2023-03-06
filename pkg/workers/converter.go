package workers

import (
	"fmt"
	config "fuu/v/pkg/config"
	"fuu/v/pkg/instrumentation"
	"fuu/v/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	pipeline = make(chan int, maxParallelizationGrade())
	quality  = "80"
)

const (
	FormatAvif string = "avif"
	FormatWebP string = "webp"
)

func Converter(workingDir string, images []string, format string, logger *zap.SugaredLogger) {
	err := os.Mkdir(filepath.Join(workingDir, format), os.ModePerm)

	if os.IsExist(err) {
		return
	}
	if err != nil && !os.IsExist(err) {
		logger.Errorw(
			"error while creating conversion directory",
			"error", err,
		)
	}

	start := time.Now()
	logger.Infow(
		"requested images conversion",
		"path", workingDir,
		"count", len(images),
		"format", format,
		"cores", maxParallelizationGrade(),
	)

	wg := new(sync.WaitGroup)
	wg.Add(len(images))

	for _, image := range images {
		pipeline <- 1
		go func(img string) {
			if utils.IsImagePath(img) {
				out := img[:len(img)-len(filepath.Ext(img))]
				cmd := exec.Command(
					"convert", filepath.Join(workingDir, img),
					"-format", format,
					"-quality", quality,
					filepath.Join(workingDir, format, fmt.Sprint(out, ".", format)),
				)
				cmd.Start()
				cmd.Wait()
			}
			<-pipeline
			wg.Done()
			instrumentation.OpsCounter.Add(1)
		}(image) // trim extension
	}

	wg.Wait()

	stop := time.Since(start)
	logger.Infow(
		"completed images conversion",
		"path", workingDir,
		"count", len(images),
		"format", format,
		"elapsed", stop,
	)
	instrumentation.TimePerOpGuage.Set(float64(stop / 1_000_000))
}

func maxParallelizationGrade() int {
	cores := runtime.NumCPU()
	format := config.Instance().ImageOptimizationFormat
	if cores == 1 {
		return 1
	}
	if cores <= 2 && format == FormatAvif {
		return 1
	}
	if cores <= 2 && format == FormatWebP {
		return 2
	}
	if cores > 2 && format == FormatAvif {
		return 1
	}
	return cores
}
