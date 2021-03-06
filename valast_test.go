package valast

import (
	"testing"
	"unsafe"

	"github.com/hexops/autogold"
	"github.com/hexops/valast/internal/test"
)

type foo struct {
	bar string
}

type baz struct {
	Bam  complex64
	zeta foo
	Beta interface{}
}

type ExportedBaz struct {
	Bam  complex64
	zeta foo
	Beta interface{}
}

func TestString(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		opt   *Options
	}{
		{
			name:  "bool",
			input: true,
		},
		{
			name:  "bool_unqualify",
			input: false,
			opt:   &Options{Unqualify: true},
		},
		{
			name:  "int32",
			input: int32(1234),
		},
		{
			name:  "int32_unqualify",
			input: int32(1234),
			opt:   &Options{Unqualify: true},
		},
		{
			name:  "uintptr",
			input: uintptr(1234),
		},
		{
			name:  "uintptr_unqualify",
			input: uintptr(1234),
			opt:   &Options{Unqualify: true},
		},
		{
			name:  "float64",
			input: float64(1.234),
		},
		{
			name:  "float64_unqualify",
			input: float64(1.234),
			opt:   &Options{Unqualify: true},
		},
		{
			name:  "complex64",
			input: complex64(1.234),
		},
		{
			name:  "complex64_unqualify",
			input: complex64(1.234),
			opt:   &Options{Unqualify: true},
		},
		{
			name:  "string",
			input: string("hello \t world"),
		},
		{
			name: "string_unqualify",
			input: string(`one
two
three`),
			opt: &Options{Unqualify: true},
		},
		{
			name: "struct_anonymous",
			input: struct {
				a, b int
				V    string
			}{a: 1, b: 2, V: "efg"},
		},
		{
			name: "struct_same_package",
			input: baz{
				Bam: 1.34,
				zeta: foo{
					bar: "hello",
				},
			},
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "struct_external_package",
			input: test.NewBaz(),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name: "array",
			input: [2]*baz{
				&baz{Beta: "foo"},
				&baz{Beta: 123},
			},
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name: "slice",
			input: []*baz{
				&baz{Beta: "foo"},
				&baz{Beta: 123},
				&baz{Beta: 3},
			},
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "nil",
			input: nil,
		},
		{
			name: "interface_anonymous_nil",
			input: interface {
				a() string
			}(nil),
		},
		{
			name: "interface_anonymous",
			input: &struct {
				v interface {
					String() string
					Baz() (err error)
				}
			}{v: test.NewBaz()},
		},
		{
			name: "interface_builtin",
			input: &struct {
				v error
			}{v: nil},
		},
		{
			name: "interface",
			input: &struct {
				v test.Bazer
			}{v: test.NewBaz()},
		},
		{
			name:  "unsafe_pointer",
			input: unsafe.Pointer(uintptr(0xdeadbeef)),
		},
		{
			name: "map",
			input: map[string]int32{
				"foo": 32,
				"bar": 64,
			},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := StringWithOptions(tst.input, tst.opt)
			autogold.Equal(t, got)
		})
	}
}

// TestRecursion tests how recursive and cyclic data types/values are handled.
func TestRecursion(t *testing.T) {
	type foo struct {
		name string
		bar  *foo
	}
	cyclic := &foo{name: "one"}
	cyclic.bar = cyclic
	tests := []struct {
		name  string
		input interface{}
		opt   *Options
	}{
		{
			name: "basic",
			input: &foo{
				name: "one",
				bar: &foo{
					name: "two",
					bar: &foo{
						name: "three",
						bar: &foo{
							name: "four",
							bar: &foo{
								name: "five",
							},
						},
					},
				},
			},
		},
		{
			name:  "struct_cyclic",
			input: cyclic,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := StringWithOptions(tst.input, tst.opt)
			autogold.Equal(t, got)
		})
	}
}

// TestEdgeCases tests known edge-cases and past bugs that do not fit any of the broader test
// categories.
func TestEdgeCases(t *testing.T) {
	var (
		nilInterfacePointerBug test.Bazer
		bazer                  test.Bazer = test.NewBaz()
	)
	tests := []struct {
		name  string
		input interface{}
		opt   *Options
	}{
		{
			name: "interface_pointer",
			input: &struct {
				v *test.Bazer
			}{v: &bazer},
		},
		{
			// Ensures it does not produce &nil:
			//
			// 	./valast_test.go:179:9: cannot take the address of nil
			// 	./valast_test.go:179:9: use of untyped nil
			//
			name: "nil_interface_pointer_bug",
			input: &struct {
				v *test.Bazer
			}{v: &nilInterfacePointerBug},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := StringWithOptions(tst.input, tst.opt)
			autogold.Equal(t, got)
		})
	}
}

// TestExportedOnly_input tests the behavior of Options.ExportedOnly when enabled with a direct unexported input.
func TestExportedOnly_input(t *testing.T) {
	type (
		unexportedBool          bool
		unexportedInt           int
		unexportedInt8          int8
		unexportedInt16         int16
		unexportedInt32         int32
		unexportedInt64         int64
		unexportedUint          uint
		unexportedUint8         uint8
		unexportedUint16        uint16
		unexportedUint32        uint32
		unexportedUint64        uint64
		unexportedUintptr       uintptr
		unexportedFloat32       float32
		unexportedFloat64       float64
		unexportedComplex64     complex64
		unexportedComplex128    complex128
		unexportedArray         [1]float32
		unexportedInterface     error
		unexportedMap           map[string]string
		unexportedPointer       *int
		unexportedSlice         []int
		unexportedString        string
		unexportedStruct        struct{ A string }
		unexportedUnsafePointer unsafe.Pointer
	)
	tests := []struct {
		name  string
		input interface{}
		opt   *Options
	}{
		{
			name: "struct_same_package",
			input: baz{
				Bam: 1.34,
				zeta: foo{
					bar: "hello",
				},
			},
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			name:  "bool",
			input: unexportedBool(true),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "int",
			input: unexportedInt(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "int8",
			input: unexportedInt8(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "int16",
			input: unexportedInt16(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "int32",
			input: unexportedInt32(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "int64",
			input: unexportedInt64(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "uint",
			input: unexportedUint(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "uint8",
			input: unexportedUint8(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "uint16",
			input: unexportedUint16(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "uint32",
			input: unexportedUint32(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "uint64",
			input: unexportedUint64(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "uintptr",
			input: unexportedUintptr(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "float32",
			input: unexportedFloat32(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "float64",
			input: unexportedFloat64(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "complex64",
			input: unexportedComplex64(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "complex128",
			input: unexportedComplex128(1),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "array",
			input: unexportedArray{1.0},
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name: "interface",
			input: struct {
				V unexportedInterface
			}{V: nil},
			opt: &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name: "map",
			input: unexportedMap{
				"a": "b",
			},
			opt: &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "map_unexported_key_type",
			input: map[unexportedInt]string{},
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "map_unexported_value_type",
			input: map[string]unexportedInt{},
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name: "map_unexported_key_omitted",
			input: map[string]interface{}{
				"a": unexportedInt(1),
			},
			opt: &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name: "map_unexported_value_omitted",
			input: map[interface{}]string{
				unexportedInt(1): "a",
			},
			opt: &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "pointer",
			input: unexportedPointer(nil),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "slice",
			input: unexportedSlice{1, 2, 3},
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "string",
			input: unexportedString("hello"),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "struct",
			input: unexportedStruct{A: "b"},
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name:  "unsafe_pointer",
			input: unexportedUnsafePointer(uintptr(0xdeadbeef)),
			opt:   &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := StringWithOptions(tst.input, tst.opt)
			autogold.Equal(t, got)
		})
	}
}

// TestExportedOnly_nested tests the behavior of Options.ExportedOnly when enabled with an unexported
// value/type nested below an exported one.
func TestExportedOnly_nested(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		opt   *Options
	}{
		{
			name:  "external_struct_unexported_field_omitted",
			input: test.NewBaz(),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			name: "struct_same_package_unexported_field_omitted",
			input: ExportedBaz{
				Bam: 1.34,
				zeta: foo{
					bar: "hello",
				},
			},
			opt: &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
		{
			name: "interface",
			input: struct {
				zeta foo
			}{zeta: foo{bar: "baz"}},
			opt: &Options{PackageName: "other", PackagePath: "github.com/other/other", ExportedOnly: true},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := StringWithOptions(tst.input, tst.opt)
			autogold.Equal(t, got)
		})
	}
}

// TestUnexportedInputs tests the behavior of Options.ExportedOnly when disabled.
func TestUnexportedInputs(t *testing.T) {
	type (
		unexportedBool          bool
		unexportedInt           int
		unexportedInt8          int8
		unexportedInt16         int16
		unexportedInt32         int32
		unexportedInt64         int64
		unexportedUint          uint
		unexportedUint8         uint8
		unexportedUint16        uint16
		unexportedUint32        uint32
		unexportedUint64        uint64
		unexportedUintptr       uintptr
		unexportedFloat32       float32
		unexportedFloat64       float64
		unexportedComplex64     complex64
		unexportedComplex128    complex128
		unexportedArray         [1]float32
		unexportedInterface     error
		unexportedMap           map[string]string
		unexportedPointer       *int
		unexportedSlice         []int
		unexportedString        string
		unexportedStruct        struct{ A string }
		unexportedUnsafePointer unsafe.Pointer
	)
	tests := []struct {
		name  string
		input interface{}
		opt   *Options
	}{
		{
			name:  "bool",
			input: unexportedBool(true),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "int",
			input: unexportedInt(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "int8",
			input: unexportedInt8(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "int16",
			input: unexportedInt16(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "int32",
			input: unexportedInt32(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "int64",
			input: unexportedInt64(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "uint",
			input: unexportedUint(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "uint8",
			input: unexportedUint8(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "uint16",
			input: unexportedUint16(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "uint32",
			input: unexportedUint32(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "uint64",
			input: unexportedUint64(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "uintptr",
			input: unexportedUintptr(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "float32",
			input: unexportedFloat32(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "float64",
			input: unexportedFloat64(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "complex64",
			input: unexportedComplex64(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "complex128",
			input: unexportedComplex128(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "array",
			input: unexportedArray{1.0},
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name: "interface",
			input: struct {
				V unexportedInterface
			}{V: nil},
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name: "map",
			input: unexportedMap{
				"a": "b",
			},
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "pointer",
			input: unexportedPointer(nil),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "slice",
			input: unexportedSlice{1, 2, 3},
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "string",
			input: unexportedString("hello"),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "struct",
			input: unexportedStruct{A: "b"},
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "unsafe_pointer",
			input: unexportedUnsafePointer(uintptr(0xdeadbeef)),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := StringWithOptions(tst.input, tst.opt)
			autogold.Equal(t, got)
		})
	}
}

func TestStringFormatting(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "short_single_line",
			input: "hello world",
		},
		{
			name:  "long_single_line_180",
			input: "hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world ",
		},
		{
			name:  "short_multi_line",
			input: "hello\nworld",
		},
		{
			name: "long_multi_line_180",
			input: `hello world hello world hello world hello world hello world hello world hello world hello world
hello world hello world hello world hello world hello world hello world hello world`,
		},
		{
			name:  "long_multi_line_with_backticks",
			input: "hello world hello world hello world hello world hello world hello world hello world hello world\n`hello world hello world hello world hello world hello world hello world hello world",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := String(tst.input)
			autogold.Equal(t, got)
		})
	}
}
