package tracing

import (
	"context"
	"testing"
)

func TestSpan(t *testing.T) {
	t.Skip()
	tracingEnabled = true
	expected := "TestSpan"

	var got string
	spy := func(o *SpanOptions) {
		got = o.label
	}
	Span(context.Background(), spy)

	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestStartSpan(t *testing.T) {
	tracingEnabled = true

	tests := []struct {
		name               string
		label              string
		funcs              []func(*SpanOptions)
		expectedLabel      string
		expectedStringTags map[string]string
		expectedIntTags    map[string]int64
	}{
		{
			name:          "the span should have the right label",
			label:         "Something",
			expectedLabel: "Something",
		},
		{
			name:          "the span should handle the namespace correctly",
			label:         "Something",
			expectedLabel: "test::Something",
			funcs: []func(*SpanOptions){
				Namespace("test"),
			},
		},
		{
			name:          "the span should handle the string tags correctly",
			label:         "Something",
			expectedLabel: "Something",
			funcs: []func(*SpanOptions){
				StringTags("foo", "bar"),
			},
			expectedStringTags: map[string]string{"foo": "bar"},
		},
		{
			name:          "the span should handle the int tags correctly",
			label:         "Something",
			expectedLabel: "Something",
			funcs: []func(*SpanOptions){
				Int64Tags("leet", int64(1337)),
			},
			expectedIntTags: map[string]int64{"leet": int64(1337)},
		},
	}

	spy := func(s *string, st *map[string]string, it *map[string]int64) func(*SpanOptions) {
		return func(o *SpanOptions) {
			*s = o.spanLabel()
			*st = o.StringTags
			*it = o.Int64Tags
		}
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			var got string
			st := make(map[string]string)
			it := make(map[string]int64)
			// spy *must* be injected last to spy correctly
			funcs := append(test.funcs, spy(&got, &st, &it))
			StartSpan(context.Background(), test.label, funcs...)
			if got != test.expectedLabel {
				t.Fatalf("expected %s, got %s", test.expectedLabel, got)
			}
			for k, v := range test.expectedStringTags {
				tag := st[k]
				if tag != v {
					t.Fatalf("expected %s, got %s", v, tag)
				}
			}
			for k, v := range test.expectedIntTags {
				tag := it[k]
				if tag != v {
					t.Fatalf("expected %d, got %d", v, tag)
				}
			}
		})
	}

}
