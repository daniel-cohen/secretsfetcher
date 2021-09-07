package logging

import (
	"fmt"

	awslog "github.com/aws/smithy-go/logging"
	"go.uber.org/zap"
)

// AwsLogger- Is a zap wrapper for use with aws-sdk-go-v2
type AwsLogger struct {
	// implementing awslog.Logger
	zl *zap.Logger
}

func NewAwsLogger(zl *zap.Logger) *AwsLogger {
	return &AwsLogger{
		zl: zap.New(zl.Core(), zap.AddCaller(), zap.AddCallerSkip(1)),
	}
}

// Logf is expected to support the standard fmt package "verbs".
// AWS onlt supoorts two levels:Warn/Debug
func (l *AwsLogger) Logf(classification awslog.Classification, format string, v ...interface{}) {
	switch classification {
	case awslog.Warn:
		l.zl.Warn(fmt.Sprintf(format, v...))
	case awslog.Debug:
		l.zl.Debug(fmt.Sprintf(format, v...))
	}
	// Otherwise, do nothing with unsupported levels.
}
