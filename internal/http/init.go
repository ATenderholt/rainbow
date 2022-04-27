package http

import (
	"github.com/ATenderholt/rainbow/logging"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func init() {
	logger = logging.NewLogger().Named("http")
}
