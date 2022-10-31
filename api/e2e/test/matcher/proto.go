package matcher

import (
	"fmt"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func ProtoEqual(expected proto.Message) types.GomegaMatcher {
	return &ProtoEqualMatcher{
		Expected: expected,
	}
}

type ProtoEqualMatcher struct {
	Expected proto.Message
}

func (matcher *ProtoEqualMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil && matcher.Expected == nil {
		return false, fmt.Errorf("Refusing to compare <nil> to <nil>.\nBe explicit and use BeNil() instead.  " +
			"This is to avoid mistakes where both sides of an assertion are erroneously uninitialized.")
	}

	message, ok := actual.(proto.Message)
	if !ok {
		return false, fmt.Errorf("invalid type, expected proto.Message, got: %t", actual)
	}

	return proto.Equal(message, matcher.Expected), nil
}

func (matcher *ProtoEqualMatcher) FailureMessage(actual interface{}) string {
	actualMessage, _ := actual.(proto.Message)
	return format.MessageWithDiff(prototext.Format(actualMessage), "to equal", prototext.Format(matcher.Expected))
}

func (matcher *ProtoEqualMatcher) NegatedFailureMessage(actual interface{}) string {
	actualMessage, _ := actual.(proto.Message)
	return format.MessageWithDiff(prototext.Format(actualMessage), "not to equal", prototext.Format(matcher.Expected))
}
