package initialize

import (
	"context"
	"time"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/task"
	"github.com/zkep/mygeektime/lib/schedule"
)

func Tw(ctx context.Context) error {

	global.TW = schedule.NewTimerWheel(200*time.Millisecond, 1000)

	global.TW.RepeatedTimer(time.Second*30, func(t time.Time) { _ = task.TaskHandler(ctx, t) }, nil)

	return nil
}
