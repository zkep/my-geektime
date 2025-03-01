package initialize

import (
	"context"
	"time"

	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/task"
	"github.com/zkep/my-geektime/lib/schedule"
)

func Tw(ctx context.Context) error {

	global.TW = schedule.NewTimerWheel(200*time.Millisecond, 1000)

	global.TW.RepeatedTimer(time.Second*10, func(t time.Time) { _ = task.TaskHandler(ctx, t) }, nil)

	return nil
}
