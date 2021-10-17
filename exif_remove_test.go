package exifremove

import (
	"io/ioutil"
	"reflect"
	"testing"

	jpegstructure "github.com/dsoprea/go-jpeg-image-structure"
	pngstructure "github.com/dsoprea/go-png-image-structure"
	"github.com/stretchr/testify/assert"
)

func TestRemoveLogic(t *testing.T) {

	data, err := ioutil.ReadFile("exif-remove-tool/img/jpg/11-tests.jpg")
	assert.Nil(t, err)
	filtered, err := Remove(data)
	assert.Nil(t, err)
	jmp := jpegstructure.NewJpegMediaParser()
	sl, err := jmp.ParseBytes(filtered)
	_, _, err = sl.Exif()
	assert.NotNil(t, err)

	data, err = ioutil.ReadFile("exif-remove-tool/img/png/exif.png")
	assert.Nil(t, err)
	filtered, err = Remove(data)
	assert.Nil(t, err)
	pmp := pngstructure.NewPngMediaParser()
	cs, err := pmp.ParseBytes(filtered)
	_, _, err = cs.Exif()
	assert.NotNil(t, err)

}

func TestRemove(t *testing.T) {
	// Define test cases
	type Args struct {
		FilePath string
	}
	type DiffType int
	//goland:noinspection GoSnakeCaseUsage
	const (
		DiffType_IsNil     DiffType = 1 << 0
		DiffType_IsNonNil  DiffType = 0
		DiffType_Equals    DiffType = 1 << 1
		DiffType_NotEquals DiffType = 0
	)
	type Results struct {
		BytesDiff DiffType
		Error     error
	}
	type TestCase struct {
		Name    string
		Args    Args
		Results Results
	}

	testCases := []TestCase{
		{Name: "exif jpg", Args: Args{FilePath: "exif-remove-tool/img/jpg/11-tests.jpg"}, Results: Results{BytesDiff: DiffType_IsNonNil | DiffType_NotEquals, Error: nil}},
		{Name: "no exif jpg", Args: Args{FilePath: "exif-remove-tool/edge_case/no_exif.jpg"}, Results: Results{BytesDiff: DiffType_IsNil | DiffType_NotEquals, Error: ErrNoExif}},
		{Name: "exif png", Args: Args{FilePath: "exif-remove-tool/img/png/exif.png"}, Results: Results{BytesDiff: DiffType_IsNonNil | DiffType_NotEquals, Error: nil}},
		{Name: "no exif png", Args: Args{FilePath: "exif-remove-tool/edge_case/no_exif.png"}, Results: Results{BytesDiff: DiffType_IsNil | DiffType_NotEquals, Error: ErrNoExif}},
		{Name: "not compatible", Args: Args{FilePath: "exif-remove-tool/edge_case/test.txt"}, Results: Results{BytesDiff: DiffType_IsNil | DiffType_NotEquals, Error: ErrNotCompatible}},
	}

	// Run tests
	t.Parallel()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			filePath := testCase.Args.FilePath
			expectedErr := testCase.Results.Error
			expectedEquals := testCase.Results.BytesDiff & DiffType_Equals == DiffType_Equals
			expectedNil := testCase.Results.BytesDiff & DiffType_IsNil == DiffType_IsNil

			// set up
			beforeBytes, actualErr := ioutil.ReadFile(filePath)
			if actualErr != nil {
				t.Fatal("ioutil.ReadFile(): can't read file: ", actualErr)
				return
			}

			// test
			afterBytes, actualErr := Remove(beforeBytes)
			if expectedErr != actualErr {
				t.Errorf("Remove(): error is not expected. Expected: %+v, Actual: %+v", expectedErr, actualErr)
				return
			}

			actualNil := afterBytes == nil
			if expectedNil != actualNil {
				t.Errorf("Remove(): afterBytes is not expected nullish. Expected: %t, Actual: %t", expectedNil, actualNil)
			}

			actualEquals := reflect.DeepEqual(beforeBytes, afterBytes)
			if expectedEquals != actualEquals {
				t.Errorf("Remove(): afterBytes is not expected equality. Expected: %t, Actual: %t", expectedEquals, actualEquals)
			}
		})
	}
}
