/*
Package internal contains code for the InterMediateDataStructure.
NOTE: these are unfinished experimental components, expect code-duplication.
*/
package internal

type ModelType int

const (
	invalidModel ModelType = iota

	// (https://wiki.theory.org/BitTorrentSpecification)
	bencodingString
	bencodingInteger
	bencodingList
	bencodingDictionary

	// (https://www.json.org/json-en.html)
	jsonObject
	jsonArray
	jsonString
	jsonNumber
	jsonNull
	jsonBool
)

type TokenType int

const (
	TokenTypeValue TokenType = iota
	TokenTypeKey
	TokenTypeListStart
	TokenTypeListEnd
	TokenTypeMapStart
	TokenTypeMapEnd
)

// InBetweener is an intermediate data structure that facilitates
// transformations between different data formats.
//
// This is analogous to the rust Serde data model,
// except that uses compile-time macro-based code-gen
// (leading to zero-cost abstractions),
// Whereas a parser implemented in Go has to rely on
// reflection to inspecting type information at runtime.
type InBetweener struct {
	Token  TokenType
	Model  ModelType
	Data   []byte
	result any // NOTE: this must be tested for a concrete type, e.g. s, ok := InBetweener.result.(string)
}

type Data struct{}

// IDEA: "compact form" that uses bit manipulation to reuse the same memory address space
