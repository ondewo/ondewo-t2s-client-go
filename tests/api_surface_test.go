// Copyright 2020-2026 ONDEWO GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Everything in this file names the ONDEWO T2S API specifically: its service, its messages, one of
// its enums. It is the ONLY file that differs from the sibling go clients - generated_code_test.go
// and auth_test.go are product agnostic and are copied over unchanged.
package tests

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	t2s "github.com/ondewo/ondewo-t2s-client-go/api/ondewo/t2s"
)

// protoFileCount is the number of .proto files below ondewo-t2s-api/ondewo that the compiler
// consumed - T2S ships a single service proto, ondewo/t2s/text-to-speech.proto. Every one of them
// has to end up in the global descriptor registry when this package is linked; a proto that
// silently stopped being compiled is otherwise invisible until a consumer misses a type.
const protoFileCount = 1

// services is every gRPC service this product exposes, keyed by the fully qualified proto name
// the ServiceDesc must declare.
var services = map[string]*grpc.ServiceDesc{
	"ondewo.t2s.Text2Speech": &t2s.Text2Speech_ServiceDesc,
}

// clientConstructors is the generated New<Service>Client of every service above. A client SDK
// that compiles but whose constructors are missing is useless, and the two generators that
// produce them (protoc-gen-go, protoc-gen-go-grpc) can disagree - so both halves are listed.
var clientConstructors = map[string]func(grpc.ClientConnInterface) any{
	"ondewo.t2s.Text2Speech": func(cc grpc.ClientConnInterface) any { return t2s.NewText2SpeechClient(cc) },
}

// expectedMethods pins RPCs by name. The descriptor cross-check in generated_code_test.go proves
// the two generators agree with each other; it cannot notice an RPC that was renamed upstream,
// because both halves would be renamed together. These are spelled out so that a rename is a
// failing test rather than a silently broken consumer.
//
// The names are the ones in the .proto - `GetT2sPipeline`, not the `GetT2SPipeline` that
// protoc-gen-go-grpc capitalises the go method to. Both halves of the generated code carry the
// proto spelling in the ServiceDesc, which is what the wire uses.
var expectedMethods = map[string][]string{
	// StreamingSynthesize is bidirectional, so it lives in ServiceDesc.Streams, not .Methods -
	// the lookup has to consider both.
	"ondewo.t2s.Text2Speech": {
		"Synthesize",
		"BatchSynthesize",
		"StreamingSynthesize",
		"NormalizeText",
		"GetT2sPipeline",
		"CreateT2sPipeline",
		"UpdateT2sPipeline",
		"DeleteT2sPipeline",
		"ListT2sPipelines",
		"ListT2sLanguages",
		"ListT2sDomains",
		"GetServiceInfo",
		"GetCustomPhonemizer",
		"CreateCustomPhonemizer",
		"UpdateCustomPhonemizer",
		"DeleteCustomPhonemizer",
		"ListCustomPhonemizer",
	},
}

// TestMessageRoundTripsThroughTheWire is the core assertion about generated message code: a value
// built in go, serialized and parsed back is the same value. The message chosen covers every field
// kind the T2S generator has to get right at once - a string scalar, a nested message, two enums
// carried inside oneofs, a proto3 `optional` scalar and a well-known google.protobuf.Struct - so a
// generator that mis-numbers a field or loses a nested type fails here.
func TestMessageRoundTripsThroughTheWire(t *testing.T) {
	t.Parallel()

	// A Struct carries the service credentials of a cloud T2S provider. Built from the fixed
	// reference instant so the value compares exactly instead of racing the clock.
	serviceConfig, err := structpb.NewStruct(map[string]any{
		"api_key":   "8f14e45fceea167a5a36dedd4bea2543",
		"region":    "eu-central-1",
		"issued_at": referenceTime.Format("2006-01-02T15:04:05Z07:00"),
		"retries":   float64(3),
	})
	if err != nil {
		t.Fatalf("structpb.NewStruct failed: %v", err)
	}

	original := &t2s.SynthesizeRequest{
		Text: "Guten Morgen, wie geht es Ihnen?",
		Config: &t2s.RequestConfig{
			T2SPipelineId:     "default_german_1",
			OneofLengthScale:  &t2s.RequestConfig_LengthScale{LengthScale: 1.15},
			Oneof_Pcm:         &t2s.RequestConfig_Pcm{Pcm: t2s.Pcm_PCM_24},
			Oneof_AudioFormat: &t2s.RequestConfig_AudioFormat{AudioFormat: t2s.AudioFormat_mp3},
			T2SServiceConfig:  serviceConfig,
			Instruction:       proto.String("a calm female voice with a slight British accent"),
		},
	}

	wire, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("proto.Marshal(%T) failed: %v", original, err)
	}
	if len(wire) == 0 {
		t.Fatal("proto.Marshal produced 0 bytes for a fully populated message")
	}

	parsed := &t2s.SynthesizeRequest{}
	if err := proto.Unmarshal(wire, parsed); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	if !proto.Equal(original, parsed) {
		t.Fatalf("round trip changed the message:\n original = %v\n  parsed = %v", original, parsed)
	}
	if got, want := parsed.GetConfig().GetAudioFormat(), t2s.AudioFormat_mp3; got != want {
		t.Errorf("enum inside a oneof after round trip = %v, want %v", got, want)
	}
	if got, want := parsed.GetConfig().GetLengthScale(), float32(1.15); got != want {
		t.Errorf("oneof member after round trip = %v, want %v", got, want)
	}
	if got, want := parsed.GetConfig().GetT2SServiceConfig().GetFields()["region"].GetStringValue(), "eu-central-1"; got != want {
		t.Errorf("google.protobuf.Struct field after round trip = %q, want %q", got, want)
	}
}

// TestRepeatedNestedMessagesRoundTrip covers the shape the phonemizer RPCs exchange: a message
// holding a repeated field of another generated message type. A generator that loses a nested type
// or mis-numbers a repeated field fails here rather than in a consumer reading an empty list.
func TestRepeatedNestedMessagesRoundTrip(t *testing.T) {
	t.Parallel()

	original := &t2s.CustomPhonemizerProto{
		Id: "german_custom_1",
		Maps: []*t2s.Map{
			{Word: "ONDEWO", PhonemeGroups: "ˈɔndeˌvoː"},
			{Word: "Wien", PhonemeGroups: "viːn"},
		},
	}

	wire, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("proto.Marshal(%T) failed: %v", original, err)
	}

	parsed := &t2s.CustomPhonemizerProto{}
	if err := proto.Unmarshal(wire, parsed); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	if !proto.Equal(original, parsed) {
		t.Fatalf("round trip changed the message:\n original = %v\n  parsed = %v", original, parsed)
	}
	if got, want := len(parsed.GetMaps()), 2; got != want {
		t.Fatalf("repeated message has %d entries after the round trip, want %d", got, want)
	}
	if got, want := parsed.GetMaps()[0].GetPhonemeGroups(), "ˈɔndeˌvoː"; got != want {
		t.Errorf("nested phoneme group after round trip = %q, want %q", got, want)
	}
}

// TestProto3ExplicitPresenceSurvivesTheWire guards the field kind that generators get wrong: a
// proto3 `optional` scalar has to keep the difference between "set to the zero value" and "not
// set". The angular target of the same compiler lost exactly this distinction, which made a
// false/0/"" unsendable; protoc-gen-go models it as a pointer, and this asserts it stays that way.
// T2S declares 11 such fields; RequestConfig.instruction is one of them.
func TestProto3ExplicitPresenceSurvivesTheWire(t *testing.T) {
	t.Parallel()

	t.Run("zero value set explicitly is transmitted", func(t *testing.T) {
		t.Parallel()

		wire, err := proto.Marshal(&t2s.RequestConfig{Instruction: proto.String("")})
		if err != nil {
			t.Fatalf("proto.Marshal failed: %v", err)
		}
		if len(wire) == 0 {
			t.Fatal("an explicitly set zero value was dropped from the wire - proto3 presence is lost")
		}

		parsed := &t2s.RequestConfig{}
		if err := proto.Unmarshal(wire, parsed); err != nil {
			t.Fatalf("proto.Unmarshal failed: %v", err)
		}
		if parsed.Instruction == nil {
			t.Fatal("Instruction is nil after the round trip, want a pointer to the empty string")
		}
		if got := *parsed.Instruction; got != "" {
			t.Errorf("Instruction = %q, want the empty string", got)
		}
	})

	t.Run("unset stays unset", func(t *testing.T) {
		t.Parallel()

		wire, err := proto.Marshal(&t2s.RequestConfig{T2SPipelineId: "unset"})
		if err != nil {
			t.Fatalf("proto.Marshal failed: %v", err)
		}

		parsed := &t2s.RequestConfig{}
		if err := proto.Unmarshal(wire, parsed); err != nil {
			t.Fatalf("proto.Unmarshal failed: %v", err)
		}
		if parsed.Instruction != nil {
			t.Errorf("Instruction = %q after a round trip that never set it, want nil", *parsed.Instruction)
		}
	})
}

// TestEnumZeroValueIsPinned pins the member at 0 and the name maps generated beside it. T2S has no
// *_UNSPECIFIED enum at all: AudioFormat's zero member is `wav`, a real choice, and Pcm's is
// PCM_16. So the assertion is that the zero value is the one the API documents, not that it
// carries a particular name - renaming or reordering the members silently changes what an unset
// request field means, which is what this catches.
func TestEnumZeroValueIsPinned(t *testing.T) {
	t.Parallel()

	var zero t2s.AudioFormat

	if zero != t2s.AudioFormat_wav {
		t.Errorf("zero value of AudioFormat = %v, want wav", zero)
	}
	if got, want := zero.String(), "wav"; got != want {
		t.Errorf("AudioFormat(0).String() = %q, want %q", got, want)
	}
	if got, want := t2s.AudioFormat_name[0], "wav"; got != want {
		t.Errorf("AudioFormat_name[0] = %q, want %q", got, want)
	}
	if got, want := t2s.AudioFormat_value["mp3"], int32(t2s.AudioFormat_mp3); got != want {
		t.Errorf("AudioFormat_value[mp3] = %d, want %d", got, want)
	}
	if got, want := int32(t2s.AudioFormat_mp3), int32(3); got != want {
		t.Errorf("AudioFormat mp3 = %d, want %d", got, want)
	}

	// The second enum of the request path, pinned the same way.
	if got, want := t2s.Pcm(0), t2s.Pcm_PCM_16; got != want {
		t.Errorf("zero value of Pcm = %v, want %v", got, want)
	}
}

// TestUnmarshalRejectsTruncatedInput asserts the generated message reports a parse error instead
// of accepting a malformed payload: field 1 (`text`) is announced as 5 bytes long but only 1
// follows.
func TestUnmarshalRejectsTruncatedInput(t *testing.T) {
	t.Parallel()

	if err := proto.Unmarshal([]byte{0x0a, 0x05, 'a'}, &t2s.SynthesizeRequest{}); err == nil {
		t.Fatal("proto.Unmarshal accepted a truncated payload, want an error")
	}
}

// text2SpeechServer is a fake ONDEWO server: it answers GetCustomPhonemizer and inherits the
// "unimplemented" behaviour of the generated base type for every other RPC of the service.
type text2SpeechServer struct {
	t2s.UnimplementedText2SpeechServer
}

func (text2SpeechServer) GetCustomPhonemizer(
	_ context.Context, req *t2s.PhonemizerId,
) (*t2s.CustomPhonemizerProto, error) {
	return &t2s.CustomPhonemizerProto{
		Id:   req.GetId(),
		Maps: []*t2s.Map{{Word: "ONDEWO", PhonemeGroups: "ˈɔndeˌvoː"}},
	}, nil
}

// TestUnaryRPCRoundTripsOverAnInProcessServer drives the generated client stub, the generated
// server stub and the generated ServiceDesc against each other over a real gRPC connection - the
// request is marshalled, routed by the method name baked into the stub, and the response is
// parsed back. Nothing here is mocked except the transport, which is in memory.
func TestUnaryRPCRoundTripsOverAnInProcessServer(t *testing.T) {
	t.Parallel()

	conn := dialInProcess(t, nil, func(srv *grpc.Server) {
		t2s.RegisterText2SpeechServer(srv, text2SpeechServer{})
	})
	client := t2s.NewText2SpeechClient(conn)

	const id = "german_custom_1"
	response, err := client.GetCustomPhonemizer(t.Context(), &t2s.PhonemizerId{Id: id})
	if err != nil {
		t.Fatalf("GetCustomPhonemizer failed: %v", err)
	}

	if got := response.GetId(); got != id {
		t.Errorf("response id = %q, want %q", got, id)
	}
	if got, want := len(response.GetMaps()), 1; got != want {
		t.Fatalf("response carries %d maps, want %d", got, want)
	}
	if got, want := response.GetMaps()[0].GetWord(), "ONDEWO"; got != want {
		t.Errorf("response map word = %q, want %q", got, want)
	}
}

// TestUnimplementedMethodIsReportedAsUnimplemented pins the other half of the generated server
// contract: an RPC the server does not implement must come back as codes.Unimplemented, not as a
// routing failure or a panic. It also proves the method is routed at all - a method missing from
// the ServiceDesc would surface as a different code.
func TestUnimplementedMethodIsReportedAsUnimplemented(t *testing.T) {
	t.Parallel()

	conn := dialInProcess(t, nil, func(srv *grpc.Server) {
		t2s.RegisterText2SpeechServer(srv, text2SpeechServer{})
	})
	client := t2s.NewText2SpeechClient(conn)

	_, err := client.DeleteCustomPhonemizer(t.Context(), &t2s.PhonemizerId{Id: "german_custom_1"})
	if got := status.Code(err); got != codes.Unimplemented {
		t.Fatalf("DeleteCustomPhonemizer returned code %v (err = %v), want %v", got, err, codes.Unimplemented)
	}
}

// TestBidiStreamingStubOpensAStream covers the one RPC the unary sweep in generated_code_test.go
// deliberately skips. Opening a stream against a server that does not implement it is answered on
// the first Recv rather than at call time, so both halves are asserted here: the generated stub
// hands out a stream, and the server reports the RPC as unimplemented over it.
func TestBidiStreamingStubOpensAStream(t *testing.T) {
	t.Parallel()

	conn := dialInProcess(t, nil, func(srv *grpc.Server) {
		t2s.RegisterText2SpeechServer(srv, text2SpeechServer{})
	})
	client := t2s.NewText2SpeechClient(conn)

	stream, err := client.StreamingSynthesize(t.Context())
	if err != nil {
		t.Fatalf("StreamingSynthesize failed to open a stream: %v", err)
	}

	// The send may or may not fail depending on when the server's status reaches the client, so
	// the assertion is made on Recv, which always observes it.
	_ = stream.Send(&t2s.StreamingSynthesizeRequest{Text: "hallo"})

	if _, err := stream.Recv(); status.Code(err) != codes.Unimplemented {
		t.Fatalf("StreamingSynthesize Recv returned code %v (err = %v), want %v", status.Code(err), err, codes.Unimplemented)
	}
}
