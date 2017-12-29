package qframe_test

import (
	"bytes"
	"fmt"
	"github.com/tobgu/qframe"
	"github.com/tobgu/qframe/filter"
	"strings"
	"testing"
)

func assertEquals(t *testing.T, expected, actual qframe.QFrame) {
	t.Helper()
	equal, reason := expected.Equals(actual)
	if !equal {
		t.Errorf("QFrames not equal, %s.\nexpected=%s\nactual=%s", reason, expected, actual)
	}
}

func assertNotErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestQFrame_Filter(t *testing.T) {
	a := qframe.New(map[string]interface{}{
		"COL.1": []int{1, 2, 3, 4, 5},
	})

	table := []struct {
		filters  []filter.Filter
		expected qframe.QFrame
	}{
		{
			[]filter.Filter{{Column: "COL.1", Comparator: ">", Arg: 3}},
			qframe.New(map[string]interface{}{"COL.1": []int{4, 5}}),
		},
		{
			[]filter.Filter{
				{Column: "COL.1", Comparator: ">", Arg: 4},
				{Column: "COL.1", Comparator: "<", Arg: 2}},
			qframe.New(map[string]interface{}{"COL.1": []int{1, 5}}),
		},
	}

	for i, tc := range table {
		t.Run(fmt.Sprintf("Filter %d", i), func(t *testing.T) {
			b := a.Filter(tc.filters...)
			assertEquals(t, tc.expected, b)
		})
	}
}

func TestQFrame_Sort(t *testing.T) {
	a := qframe.New(map[string]interface{}{
		"COL.1": []int{0, 1, 3, 2},
		"COL.2": []int{3, 2, 1, 1},
	})

	table := []struct {
		orders   []qframe.Order
		expected qframe.QFrame
	}{
		{
			[]qframe.Order{{Column: "COL.1"}},
			qframe.New(map[string]interface{}{
				"COL.1": []int{0, 1, 2, 3},
				"COL.2": []int{3, 2, 1, 1}}),
		},
		{
			[]qframe.Order{{Column: "COL.1", Reverse: true}},
			qframe.New(map[string]interface{}{
				"COL.1": []int{3, 2, 1, 0},
				"COL.2": []int{1, 1, 2, 3}}),
		},
		{
			[]qframe.Order{{Column: "COL.2"}, {Column: "COL.1"}},
			qframe.New(map[string]interface{}{
				"COL.1": []int{2, 3, 1, 0},
				"COL.2": []int{1, 1, 2, 3}}),
		},
	}

	for i, tc := range table {
		t.Run(fmt.Sprintf("Sort %d", i), func(t *testing.T) {
			b := a.Sort(tc.orders...)
			assertEquals(t, tc.expected, b)
		})
	}
}

func TestQFrame_SortNilString(t *testing.T) {
	a, b, c := "a", "b", "c"
	in := qframe.New(map[string]interface{}{
		"COL.1": []*string{&b, nil, &a, nil, &c, &a, nil},
	})

	table := []struct {
		orders   []qframe.Order
		expected map[string]interface{}
	}{
		{
			[]qframe.Order{{Column: "COL.1"}},
			map[string]interface{}{
				"COL.1": []*string{&a, &a, &b, &c, nil, nil, nil},
			},
		},
		{
			[]qframe.Order{{Column: "COL.1", Reverse: true}},
			map[string]interface{}{
				"COL.1": []*string{nil, nil, nil, &c, &b, &a, &a},
			},
		},
	}

	for i, tc := range table {
		t.Run(fmt.Sprintf("Sort %d", i), func(t *testing.T) {
			out := in.Sort(tc.orders...)
			assertNotErr(t, out.Err)
			assertEquals(t, qframe.New(tc.expected), out)
		})
	}
}

func TestQFrame_SortStability(t *testing.T) {
	a := qframe.New(map[string]interface{}{
		"COL.1": []int{0, 1, 3, 2},
		"COL.2": []int{1, 1, 1, 1},
	})

	table := []struct {
		orders   []qframe.Order
		expected qframe.QFrame
	}{
		{
			[]qframe.Order{{Column: "COL.2", Reverse: true}, {Column: "COL.1"}},
			qframe.New(map[string]interface{}{
				"COL.1": []int{0, 1, 2, 3},
				"COL.2": []int{1, 1, 1, 1}}),
		},
	}

	for i, tc := range table {
		t.Run(fmt.Sprintf("Sort %d", i), func(t *testing.T) {
			b := a.Sort(tc.orders...)
			assertEquals(t, tc.expected, b)
		})
	}
}

func TestQFrame_Distinct(t *testing.T) {
	table := []struct {
		input    map[string]interface{}
		expected map[string]interface{}
		columns  []string
	}{
		{
			input: map[string]interface{}{
				"COL.1": []int{0, 1, 0, 1},
				"COL.2": []int{0, 1, 0, 1}},
			expected: map[string]interface{}{
				"COL.1": []int{0, 1},
				"COL.2": []int{0, 1}},
			columns: []string{"COL.1", "COL.2"},
		},
		{
			input: map[string]interface{}{
				"COL.1": []int{},
				"COL.2": []int{}},
			expected: map[string]interface{}{
				"COL.1": []int{},
				"COL.2": []int{}},
			columns: []string{"COL.1", "COL.2"},
		},
	}

	for i, tc := range table {
		t.Run(fmt.Sprintf("Distinct %d", i), func(t *testing.T) {
			in := qframe.New(tc.input)
			out := in.Distinct()
			assertEquals(t, qframe.New(tc.expected), out)
		})
	}
}

func TestQFrame_GroupByAggregate(t *testing.T) {
	table := []struct {
		input        map[string]interface{}
		expected     map[string]interface{}
		groupColumns []string
		aggregations []string
	}{
		{
			input: map[string]interface{}{
				"COL.1": []int{0, 0, 1, 2},
				"COL.2": []int{0, 0, 1, 1},
				"COL.3": []int{1, 2, 5, 7}},
			expected: map[string]interface{}{
				"COL.1": []int{0, 1, 2},
				"COL.2": []int{0, 1, 1},
				"COL.3": []int{3, 5, 7}},
			groupColumns: []string{"COL.1", "COL.2"},
			aggregations: []string{"sum", "COL.3"},
		},
		{
			input: map[string]interface{}{
				"COL.1": []int{},
				"COL.2": []int{}},
			expected: map[string]interface{}{
				"COL.1": []int{},
				"COL.2": []int{}},
			groupColumns: []string{"COL.1"},
			aggregations: []string{"sum", "COL.2"},
		},
	}

	for i, tc := range table {
		t.Run(fmt.Sprintf("GroupByAggregate %d", i), func(t *testing.T) {
			in := qframe.New(tc.input)
			out := in.GroupBy(tc.groupColumns...).Aggregate(tc.aggregations...)
			assertEquals(t, qframe.New(tc.expected), out)
		})
	}
}

func TestQFrame_Select(t *testing.T) {
	table := []struct {
		input      map[string]interface{}
		expected   map[string]interface{}
		selectCols []string
	}{
		{
			input: map[string]interface{}{
				"COL.1": []int{0, 1},
				"COL.2": []int{1, 2}},
			expected: map[string]interface{}{
				"COL.1": []int{0, 1}},
			selectCols: []string{"COL.1"},
		},
		{
			input: map[string]interface{}{
				"COL.1": []int{0, 1},
				"COL.2": []int{1, 2}},
			expected:   map[string]interface{}{},
			selectCols: []string{},
		},
	}

	for i, tc := range table {
		t.Run(fmt.Sprintf("Select %d", i), func(t *testing.T) {
			in := qframe.New(tc.input)
			out := in.Select(tc.selectCols...)
			assertEquals(t, qframe.New(tc.expected), out)
		})
	}
}

func TestQFrame_Slice(t *testing.T) {
	table := []struct {
		input    map[string]interface{}
		expected map[string]interface{}
		start    int
		end      int
	}{
		{
			input: map[string]interface{}{
				"COL.1": []float64{0.0, 1.5, 2.5, 3.5},
				"COL.2": []int{1, 2, 3, 4}},
			expected: map[string]interface{}{
				"COL.1": []float64{1.5, 2.5},
				"COL.2": []int{2, 3}},
			start: 1,
			end:   3,
		},
		{
			input: map[string]interface{}{
				"COL.1": []int{},
				"COL.2": []int{}},
			expected: map[string]interface{}{
				"COL.1": []int{},
				"COL.2": []int{}},
			start: 0,
			end:   0,
		},
	}

	for i, tc := range table {
		t.Run(fmt.Sprintf("Slice %d", i), func(t *testing.T) {
			in := qframe.New(tc.input)
			out := in.Slice(tc.start, tc.end)
			assertEquals(t, qframe.New(tc.expected), out)
		})
	}
}

func TestQCacheFrame_ReadCsv(t *testing.T) {
	/*
		Pandas reference
		>>> data = """
		... foo,bar,baz,qux
		... ccc,,,www
		... aaa,3.25,7,"""
		>>> pd.read_csv(StringIO(data))
		   foo   bar  baz  qux
		0  ccc   NaN  NaN  www
		1  aaa  3.25  7.0  NaN
	*/
	a, b, c, empty := "a", "b", "c", ""
	table := []struct {
		inputHeaders []string
		inputData    string
		emptyNull    bool
		expected     map[string]interface{}
	}{
		{
			inputHeaders: []string{"foo", "bar"},
			inputData:    "1,2\n3,4\n",
			expected: map[string]interface{}{
				"foo": []int{1, 3},
				"bar": []int{2, 4}},
		},
		{
			inputHeaders: []string{"int", "float", "bool", "string"},
			inputData:    "1,2.5,true,hello\n10,20.5,false,\"bye, bye\"",
			expected: map[string]interface{}{
				"int":    []int{1, 10},
				"float":  []float64{2.5, 20.5},
				"bool":   []bool{true, false},
				"string": []string{"hello", "bye, bye"}},
		},
		{
			inputHeaders: []string{"foo", "bar"},
			inputData:    "a,b\n,c",
			emptyNull:    true,
			expected: map[string]interface{}{
				"foo": []*string{&a, nil},
				"bar": []*string{&b, &c}},
		},
		{
			inputHeaders: []string{"foo", "bar"},
			inputData:    "a,b\n,c",
			emptyNull:    false,
			expected: map[string]interface{}{
				"foo": []*string{&a, &empty},
				"bar": []*string{&b, &c}},
		},
	}

	for i, tc := range table {
		t.Run(fmt.Sprintf("ReadCsv %d", i), func(t *testing.T) {
			input := strings.Join(tc.inputHeaders, ",") + "\n" + tc.inputData
			out := qframe.ReadCsv(strings.NewReader(input), qframe.EmptyNull(tc.emptyNull))
			assertNotErr(t, out.Err)
			assertEquals(t, qframe.New(tc.expected, qframe.ColumnOrder(tc.inputHeaders...)), out)
		})
	}
}

func TestQFrame_ReadJson(t *testing.T) {
	/*
		>>> pd.DataFrame.from_records([dict(a=1.5), dict(a=None)])
			 a
		0  1.5
		1  NaN
		>>> pd.DataFrame.from_records([dict(a=1), dict(a=None)])
			 a
		0  1.0
		1  NaN
		>>> pd.DataFrame.from_records([dict(a=1), dict(a=2)])
		   a
		0  1
		1  2
		>>> pd.DataFrame.from_records([dict(a='foo'), dict(a=None)])
			  a
		0   foo
		1  None
		>>> pd.DataFrame.from_records([dict(a=1.5), dict(a='N/A')])
			 a
		0  1.5
		1  N/A
		>>> x = pd.DataFrame.from_records([dict(a=1.5), dict(a='N/A')])
		>>> x.ix[0]
		a    1.5
		Name: 0, dtype: object
	*/
	testString := "FOO"
	table := []struct {
		input    string
		expected map[string]interface{}
	}{
		{
			input: `{"STRING1": ["a", "b"], "INT1": [1, 2], "FLOAT1": [1.5, 2.5], "BOOL1": [true, false]}`,
			expected: map[string]interface{}{
				"STRING1": []string{"a", "b"}, "INT1": []int{1, 2}, "FLOAT1": []float64{1.5, 2.5}, "BOOL1": []bool{true, false}},
		},
		{
			input:    `{"STRING1": ["FOO", null]}`,
			expected: map[string]interface{}{"STRING1": []*string{&testString, nil}},
		},
		{
			input: `[
				{"STRING1": "a", "INT1": 1, "FLOAT1": 1.5, "BOOL1": true},
				{"STRING1": "b", "INT1": 2, "FLOAT1": 2.5, "BOOL1": false}]`,
			expected: map[string]interface{}{
				// NOTE: The integers become floats if not explicitly typed
				"STRING1": []string{"a", "b"}, "INT1": []float64{1, 2}, "FLOAT1": []float64{1.5, 2.5}, "BOOL1": []bool{true, false}},
		},
		{
			input: `[{"STRING1": "FOO"}, {"STRING1": null}]`,
			expected: map[string]interface{}{
				"STRING1": []*string{&testString, nil}},
		},
	}

	for i, tc := range table {
		t.Run(fmt.Sprintf("FromJSON %d", i), func(t *testing.T) {
			out := qframe.ReadJson(strings.NewReader(tc.input))
			assertNotErr(t, out.Err)
			assertEquals(t, qframe.New(tc.expected), out)
		})
	}
}

func TestQFrame_ToCsv(t *testing.T) {
	table := []struct {
		input    map[string]interface{}
		expected string
	}{
		{
			input: map[string]interface{}{
				"STRING1": []string{"a", "b,c"}, "INT1": []int{1, 2}, "FLOAT1": []float64{1.5, 2.5}, "BOOL1": []bool{true, false}},
			expected: `BOOL1,FLOAT1,INT1,STRING1
true,1.5,1,a
false,2.5,2,"b,c"
`,
		},
	}

	for i, tc := range table {
		t.Run(fmt.Sprintf("ToCsv %d", i), func(t *testing.T) {
			in := qframe.New(tc.input)
			assertNotErr(t, in.Err)

			buf := new(bytes.Buffer)
			err := in.ToCsv(buf)
			assertNotErr(t, err)

			result := buf.String()
			if result != tc.expected {
				t.Errorf("QFrames not equal, %s ||| %s", result, tc.expected)
			}
		})
	}
}

func TestQCacheFrame_ToFromJSON(t *testing.T) {
	table := []struct {
		orientation string
	}{
		{orientation: "records"},
		{orientation: "columns"},
	}

	for i, tc := range table {
		t.Run(fmt.Sprintf("Sort %d", i), func(t *testing.T) {
			buf := new(bytes.Buffer)
			data := map[string]interface{}{
				"STRING1": []string{"añ", "bö☺	"}, "FLOAT1": []float64{1.5, 2.5}, "BOOL1": []bool{true, false}}
			originalDf := qframe.New(data)
			err := originalDf.ToJson(buf, tc.orientation)
			assertNotErr(t, err)

			jsonDf := qframe.ReadJson(buf)
			assertNotErr(t, jsonDf.Err)
			assertEquals(t, originalDf, jsonDf)
		})
	}
}