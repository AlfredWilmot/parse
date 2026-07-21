/*
Package core contains code for the InterMediateDataStructure.
NOTE: these are unfinished experimental components, expect code-duplication.
*/
package core

type InBetweenerType int

const (
	bencodingByteStringType InBetweenerType = iota
	bencodingIntegerType
	bencodingListType
	bencodingDictionaryType
)

type InBetweenerInterface interface {
	Type() InBetweenerType
	String() string
}

// InBetweener is an intermediate data structure that facilitates
// transformations between different data formats.
//
// This is analogous to the rust Serde data model,
// except that uses compile-time macro-based code-gen
// (leading to zero-cost abstractions),
// Whereas a parser implemented in Go has to rely on
// runtime reflection to inspecting type information at runtime
type InBetweener struct {
	ASCIIString          string
	Int64Val             int64
	SelfValueList        []InBetweener
	StrKeyedSelfValueMap map[string]InBetweener
}

// IDEA: "compact form" that uses bit manipulation to reuse the same memory address space
