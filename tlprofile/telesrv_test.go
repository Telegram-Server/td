package tlprofile_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/iamxvbaba/td/bin"
	"github.com/iamxvbaba/td/tg"
	"github.com/iamxvbaba/td/tlprofile"
)

func TestTelesrvMethodsDispatchInEveryProfile(t *testing.T) {
	methods := []struct {
		semantic tlprofile.SemanticID
		request  bin.Object
	}{
		{tlprofile.SemanticMethodTelesrvGetLoginPassword, &tg.TelesrvGetLoginPasswordRequest{}},
		{tlprofile.SemanticMethodTelesrvSignInWithPassword, &tg.TelesrvSignInWithPasswordRequest{Password: &tg.InputCheckPasswordEmpty{}}},
		{tlprofile.SemanticMethodTelesrvSignUpWithPassword, &tg.TelesrvSignUpWithPasswordRequest{}},
		{tlprofile.SemanticMethodTelesrvGetMyLoginPassword, &tg.TelesrvGetMyLoginPasswordRequest{}},
		{tlprofile.SemanticMethodTelesrvUpdateLoginPassword, &tg.TelesrvUpdateLoginPasswordRequest{Current: &tg.InputCheckPasswordEmpty{}}},
	}
	profiles := []tlprofile.Profile{
		tlprofile.Profile225, tlprofile.Profile226, tlprofile.Profile227,
		tlprofile.Profile228, tlprofile.Profile229, tlprofile.Profile230,
	}
	for _, method := range methods {
		for _, profile := range profiles {
			t.Run(fmt.Sprintf("%T/%d", method.request, profile), func(t *testing.T) {
				dispatcher := tlprofile.NewDispatcher()
				var got bin.Object
				if err := dispatcher.Register(method.semantic, func(_ context.Context, request bin.Object) (any, error) {
					got = request
					return &tg.BoolTrue{}, nil
				}); err != nil {
					t.Fatal(err)
				}
				var wire bin.Buffer
				if err := method.request.Encode(&wire); err != nil {
					t.Fatal(err)
				}
				admission, err := dispatcher.Admit(profile, &wire, tlprofile.Limits{})
				if err != nil {
					t.Fatal(err)
				}
				if admission.Call().Method() != method.semantic {
					t.Fatalf("method = %#x, want %#x", admission.Call().Method(), method.semantic)
				}
				_, _ = dispatcher.Dispatch(context.Background(), admission)
				if fmt.Sprintf("%T", got) != fmt.Sprintf("%T", method.request) {
					t.Fatalf("handler got %T, want %T", got, method.request)
				}
			})
		}
	}
}
