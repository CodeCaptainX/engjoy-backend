package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	apiresponse "sentenceminer/pkg/http/response"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Log level icons for visual distinction
const (
	IconFatal = "💀"
	IconError = "❌"
	IconWarn  = "⚠️ "
	IconInfo  = "ℹ️ "
	IconDebug = "🔍"
	IconTrace = "📝"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGray   = "\033[90m"
	ColorCyan   = "\033[36m"
	ColorGreen  = "\033[32m"
	ColorBold   = "\033[1m"
)

func init() {
	// Configure zerolog for human-readable console output
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			zerolog.MessageFieldName,
		},
		FormatLevel: func(i interface{}) string {
			var icon string
			if ll, ok := i.(string); ok {
				switch ll {
				case "fatal":
					icon = IconFatal
				case "error":
					icon = IconError
				case "warn":
					icon = IconWarn
				case "info":
					icon = IconInfo
				case "debug":
					icon = IconDebug
				default:
					icon = IconTrace
				}
				return fmt.Sprintf("\n %s %s : ", icon, ll)
			}
			return fmt.Sprintf("%v", i)
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("\n  %s%s%s", ColorBold, i, ColorReset)
		},
		FormatFieldName: func(i interface{}) string {
			return fmt.Sprintf("\n  %s%s:%s ", ColorCyan, i, ColorReset)
		},
		FormatFieldValue: func(i interface{}) string {
			return fmt.Sprintf("%s%s%s", ColorGreen, i, ColorReset)
		},
	}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()
}

// CustomLog represents a custom log entry with additional context.
type CustomLog struct {
	MessageID string
	LogReason string
	Function  string
	File      string
	FullPath  string
	Line      int
	Timestamp time.Time
	Level     string
}

// LogToString returns a formatted log message string with icons.
func (e *CustomLog) LogToString() string {
	shortFile := filepath.Base(e.File)
	icon := e.getLevelIcon()

	return fmt.Sprintf(
		"%s [%s] 🕐 %s | 🆔 %s | ⚙️  %s | 📄 %s:%d | 📂 %s | 💬 %s",
		icon,
		e.Level,
		e.Timestamp.Format("2006-01-02 15:04:05"),
		e.MessageID,
		e.Function,
		shortFile,
		e.Line,
		e.FullPath,
		e.LogReason,
	)
}

// getLevelIcon returns the appropriate icon for the log level.
func (e *CustomLog) getLevelIcon() string {
	switch e.Level {
	case "fatal":
		return IconFatal
	case "error":
		return IconError
	case "warn":
		return IconWarn
	case "info":
		return IconInfo
	case "debug":
		return IconDebug
	default:
		return IconTrace
	}
}

// NewCustomLog creates a new CustomLog with caller information and enhanced formatting.
func NewCustomLog(messageID string, logDesc string, logType ...string) *CustomLog {
	pc, file, line, ok := runtime.Caller(1)
	function := "unknown"
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			function = fn.Name()
		}
	}

	// Determine log level
	level := "info"
	if len(logType) > 0 {
		level = logType[0]
	}

	msg := &CustomLog{
		MessageID: messageID,
		LogReason: logDesc,
		Function:  function,
		File:      file,
		FullPath:  file,
		Line:      line,
		Timestamp: time.Now(),
		Level:     level,
	}

	// Get short filename for cleaner logs
	// shortFile := filepath.Base(file)

	// Rich, detailed message format with clickable file path
	message := fmt.Sprintf("⚙️  %s() : ",
		function,
	)

	// Log based on level with full details
	switch level {
	case "fatal":
		log.Fatal().
			Str("🆔 id", messageID).
			Str("💬 reason", logDesc).
			Str("📂 path", fmt.Sprintf("%s:%d", file, line)).
			Time("🕐 timestamp", msg.Timestamp).
			Msg(message)
	case "error":
		log.Error().
			Str("🆔 id", messageID).
			Str("💬 reason", logDesc).
			Str("📂 path", fmt.Sprintf("%s:%d", file, line)).
			Time("🕐 timestamp", msg.Timestamp).
			Msg(message)
	case "warn":
		log.Warn().
			Str("🆔 id", messageID).
			Str("💬 reason", logDesc).
			Str("📂 path", fmt.Sprintf("%s:%d", file, line)).
			Time("🕐 timestamp", msg.Timestamp).
			Msg(message)
	case "info":
		log.Info().
			Str("🆔 id", messageID).
			Str("💬 reason", logDesc).
			Str("📂 path", fmt.Sprintf("%s:%d", file, line)).
			Time("🕐 timestamp", msg.Timestamp).
			Msg(message)
	case "debug":
		log.Debug().
			Str("🆔 id", messageID).
			Str("💬 reason", logDesc).
			Str("📂 path", fmt.Sprintf("%s:%d", file, line)).
			Time("🕐 timestamp", msg.Timestamp).
			Msg(message)
	case "trace":
		log.Trace().
			Str("🆔 id", messageID).
			Str("💬 reason", logDesc).
			Str("📂 path", fmt.Sprintf("%s:%d", file, line)).
			Time("🕐 timestamp", msg.Timestamp).
			Msg(message)
	default:
		if zerolog.GlobalLevel() <= zerolog.InfoLevel {
			log.Info().
				Str("🆔 id", messageID).
				Str("💬 reason", logDesc).
				Str("📂 path", fmt.Sprintf("%s:%d", file, line)).
				Time("🕐 timestamp", msg.Timestamp).
				Msg(message)
		}
	}

	return msg
}

// LogWithContext adds additional context fields to the log entry.
func (e *CustomLog) LogWithContext(fields map[string]interface{}, logType string) {
	event := log.With().
		Str("🆔 id", e.MessageID).
		Str("⚙️  func", e.Function).
		Str("📂 path", fmt.Sprintf("%s:%d", e.FullPath, e.Line)).
		Time("🕐 timestamp", e.Timestamp).
		Fields(fields).
		Logger()

	msg := fmt.Sprintf("⚙️  %s() | 📂 %s:%d | 💬 %s ",
		e.Function,
		e.FullPath,
		e.Line,
		e.LogReason,
	)

	switch logType {
	case "fatal":
		event.Fatal().Msg(msg)
	case "error":
		event.Error().Msg(msg)
	case "warn":
		event.Warn().Msg(msg)
	case "info":
		event.Info().Msg(msg)
	case "debug":
		event.Debug().Msg(msg)
	default:
		event.Info().Msg(msg)
	}
}

// Quick logging helpers
func Fatal(messageID, message string) *CustomLog {
	return NewCustomLog(messageID, message, "fatal")
}

func Error(messageID, message string) *CustomLog {
	return NewCustomLog(messageID, message, "error")
}

func Warn(messageID, message string) *CustomLog {
	return NewCustomLog(messageID, message, "warn")
}

func Info(messageID, message string) *CustomLog {
	return NewCustomLog(messageID, message, "info")
}

func Debug(messageID, message string) *CustomLog {
	return NewCustomLog(messageID, message, "debug")
}

type ErrorParams struct {
	LogMessage string
	I18nKey    string
	Err        error
	TypeError  string
	StatusCode int
}

func NewCustomLogResponse(params ErrorParams) *apiresponse.ErrorResponse {
	if params.LogMessage == "" {
		params.LogMessage = "Default Log Message"
	}
	if params.I18nKey == "" {
		params.I18nKey = "Default I18n Key"
	}
	if params.Err == nil {
		params.Err = fmt.Errorf("default Error")
	}

	NewCustomLog(params.LogMessage, params.Err.Error(), params.TypeError)

	return (&apiresponse.ErrorResponse{}).NewErrorResponse(params.I18nKey, params.Err, params.StatusCode)
}
