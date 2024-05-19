package test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/misc"
	"github.com/tcodes0/go/src/reflectutil"
)

//nolint:funlen // test
func TestEnvTagResolve(t *testing.T) {
	assert := require.New(t)

	type dog struct {
		Name  string `default:"kim"   env:"DOG_NAME"`
		Breed string `env:"DOG_BREED"`
		Owner string
	}

	envTag := reflectutil.EnvTag{
		Tag:     "env",
		Default: "default",
	}

	tests := []struct {
		name     string
		dog      *dog
		envTag   reflectutil.EnvTag
		returns  require.ErrorAssertionFunc
		nameEnv  string
		breedEnv string
		ownerEnv string
		nameEq   string
		breedEq  string
		ownerEq  string
	}{
		{
			name:     "Sets field value",
			envTag:   envTag,
			dog:      misc.PointerTo(dog{}),
			returns:  require.NoError,
			nameEnv:  "fido",
			breedEnv: "golden",
			ownerEnv: "",
			nameEq:   "fido",
			breedEq:  "golden",
			ownerEq:  "",
		},
		{
			name:     "Defaults",
			envTag:   envTag,
			dog:      misc.PointerTo(dog{}),
			returns:  require.NoError,
			nameEnv:  "",
			breedEnv: "golden",
			ownerEnv: "",
			nameEq:   "kim",
			breedEq:  "golden",
			ownerEq:  "",
		},
		{
			name:     "No change to not-tagged",
			envTag:   envTag,
			dog:      misc.PointerTo(dog{}),
			returns:  require.NoError,
			nameEnv:  "",
			breedEnv: "golden",
			ownerEnv: "leopoldo",
			nameEq:   "kim",
			breedEq:  "golden",
			ownerEq:  "",
		},
		{
			name:     "No overwrite",
			envTag:   envTag,
			dog:      misc.PointerTo(dog{Name: "julia"}),
			returns:  require.NoError,
			nameEnv:  "fido",
			breedEnv: "golden",
			ownerEnv: "",
			nameEq:   "julia",
			breedEq:  "golden",
			ownerEq:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("DOG_NAME", test.nameEnv)
			t.Setenv("DOG_BREED", test.breedEnv)
			t.Setenv("DOG_OWNER", test.ownerEnv)
			test.returns(t,
				test.envTag.Resolve(misc.PointerTo(reflect.TypeOf(test.dog).Elem().Field(0)), reflect.ValueOf(test.dog).Elem().Field(0)),
			)
			test.returns(t,
				test.envTag.Resolve(misc.PointerTo(reflect.TypeOf(test.dog).Elem().Field(1)), reflect.ValueOf(test.dog).Elem().Field(1)),
			)
			test.returns(t,
				test.envTag.Resolve(misc.PointerTo(reflect.TypeOf(test.dog).Elem().Field(2)), reflect.ValueOf(test.dog).Elem().Field(2)),
			)
			assert.Equal(test.nameEq, test.dog.Name)
			assert.Equal(test.breedEq, test.dog.Breed)
			assert.Equal(test.ownerEq, test.dog.Owner)
		})
	}
}

// func TestEnvTag(t *testing.T) {
// func TestEnvTagDefault(t *testing.T) {
// func TestEnvTagEmpty(t *testing.T) {
// func TestEnvTagEmptyFallback(t *testing.T) {
// func TestEnvTagNoOverwrite(t *testing.T) {
// func TestEnvTagErrNotString(t *testing.T) {
// func TestEnvTagNoop(t *testing.T) {
// 	t.Parallel()
// 	assert := require.New(t)

// 	type config struct {
// 		Ptr *string
// 		Sl  []byte
// 		Ok  bool
// 	}

// 	cfg := &config{
// 		Ok:  true,
// 		Ptr: misc.PointerTo(foo),
// 		Sl:  []byte(foo),
// 	}
// 	getKey := func(string) string {
// 		return foo
// 	}
// 	expected := &config{
// 		Ok:  true,
// 		Ptr: misc.PointerTo(foo),
// 		Sl:  []byte(foo),
// 	}

// 	err := configkey.Run(cfg, getKey)
// 	assert.NoError(err)
// 	assert.Equal(expected, cfg)
// }
