package main

import (
	"flag"
	"strings"
	"time"
)

// timeFlag is a flag.Value that allows parsing a time value from RFC 3339
// format.
type timeFlag struct {
	target *time.Time
}

var _ flag.Value = timeFlag{}

func (tf timeFlag) String() string {
	if tf.target.IsZero() {
		return ""
	}
	return tf.target.Format(time.RFC3339Nano)
}

func (tf timeFlag) Set(v string) error {
	if v == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return err
	}
	*tf.target = t
	return nil
}

// stringSlice is a flag.Value that allows parsing a comma-separated strings.
type stringSlice struct {
	target *[]string
}

var _ flag.Value = stringSlice{}

func (tf stringSlice) String() string {
	if tf.target == nil {
		return ""
	}
	return strings.Join(*tf.target, ",")
}

func (tf stringSlice) Set(v string) error {
	if v == "" {
		return nil
	}
	*tf.target = strings.Split(v, ",")
	return nil
}

// nilWriter is no-operation writer.
type nilWriter struct{}

func (nilWriter) Write(data []byte) (int, error) {
	return len(data), nil
}
