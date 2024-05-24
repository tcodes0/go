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
			dog:      misc.ToPtr(dog{}),
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
			dog:      misc.ToPtr(dog{}),
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
			dog:      misc.ToPtr(dog{}),
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
			dog:      misc.ToPtr(dog{Name: "julia"}),
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

			for i := range 3 {
				test.returns(t,
					test.envTag.UpdateField(misc.ToPtr(reflect.TypeOf(test.dog).Elem().Field(i)), reflect.ValueOf(test.dog).Elem().Field(i)),
				)
			}

			assert.Equal(test.nameEq, test.dog.Name)
			assert.Equal(test.breedEq, test.dog.Breed)
			assert.Equal(test.ownerEq, test.dog.Owner)
		})
	}
}
