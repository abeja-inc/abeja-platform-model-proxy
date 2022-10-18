package logging

import (
	"bytes"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type SimpleFormatter struct {
	// TimestampFormat sets the format used for marshaling timestamps.
	TimestampFormat string

	// FieldMap allows users to customize the names of keys for default fields.
	// As an example:
	// formatter := &JSONFormatter{
	//   	FieldMap: FieldMap{
	// 		 FieldKeyTime:  "@timestamp",
	// 		 FieldKeyLevel: "@level",
	// 		 FieldKeyMsg:   "@message",
	// 		 FieldKeyFunc:  "@caller",
	//    },
	// }
	FieldMap FieldMap
}

func (f *SimpleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339
	}

	b := &bytes.Buffer{}

	timestamp := entry.Time.Format(timestampFormat)
	f.appendValue(b, timestamp)
	f.appendValue(b, entry.Message)

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *SimpleFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}

	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}
	b.WriteString(fmt.Sprintf("%#v", stringVal))
}
