package intseries

import (
	"fmt"
	"github.com/tobgu/go-qcache/dataframe/filter"
	"github.com/tobgu/go-qcache/dataframe/internal/series"
)

type IntSeries struct {
	data []int
}

func New(d []int) IntSeries {
	return IntSeries{data: d}
}

var intFilterFuncs = map[filter.Comparator]func([]uint32, []int, interface{}, []bool) error{
	filter.Gt: gt,
	filter.Lt: lt,
}

func (s IntSeries) Filter(index []uint32, c filter.Comparator, comparatee interface{}, bIndex []bool) error {
	// TODO: Also make it possible to compare to values in other column
	compFunc, ok := intFilterFuncs[c]
	if !ok {
		return fmt.Errorf("invalid comparison operator for int, %s", c)
	}

	return compFunc(index, s.data, comparatee, bIndex)
}

func (s IntSeries) Equals(index []uint32, other series.Series, otherIndex []uint32) bool {
	otherI, ok := other.(IntSeries)
	if !ok {
		return false
	}

	for ix, x := range index {
		if s.data[x] != otherI.data[otherIndex[ix]] {
			return false
		}
	}

	return true
}

func (s IntSeries) Subset(index []uint32) series.Series {
	data := make([]int, 0, len(index))
	for _, ix := range index {
		data = append(data, s.data[ix])
	}

	return IntSeries{data: data}
}

func (s IntSeries) Comparable(reverse bool) series.Comparable {
	if reverse {
		return IntComparable{data: s.data, ltValue: series.GreaterThan, gtValue: series.LessThan}
	}

	return IntComparable{data: s.data, ltValue: series.LessThan, gtValue: series.GreaterThan}
}

type IntComparable struct {
	data []int
	ltValue series.CompareResult
	gtValue series.CompareResult
}

func (c IntComparable) Compare(i, j uint32) series.CompareResult {
	x, y := c.data[i], c.data[j]
	if x < y {
		return c.ltValue
	}

	if x > y {
		return c.gtValue
	}

	return series.Equal
}

// TODO: Some kind of code generation for all the below functions for all supported types

func gt(index []uint32, column []int, comparatee interface{}, bIndex []bool) error {
	comp, ok := comparatee.(int)
	if !ok {
		return fmt.Errorf("invalid comparison type")
	}

	for i, x := range bIndex {
		bIndex[i] = x || column[index[i]] > comp
	}

	return nil
}

func lt(index []uint32, column []int, comparatee interface{}, bIndex []bool) error {
	comp, ok := comparatee.(int)
	if !ok {
		return fmt.Errorf("invalid comparison type")
	}

	for i, x := range bIndex {
		bIndex[i] = x || column[index[i]] < comp
	}

	return nil
}
