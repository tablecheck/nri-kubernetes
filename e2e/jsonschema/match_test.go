package jsonschema

import (
	"testing"

	"os"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
)

var s = EventTypeToSchemaFilepath{
	"TestNodeSample":    "testdata/schema-testnode.json",
	"TestServiceSample": "testdata/schema-testservice.json",
}

func TestNoError(t *testing.T) {
	c := readTestInput(t, "testdata/input-complete.json")

	err := Match(c, s)
	assert.NoError(t, err)
}

func TestErrorValidatingInputWithNoData(t *testing.T) {
	c := readTestInput(t, "testdata/input-invalid-nodata.json")

	err := Match(c, s)
	assert.Contains(t, err.Error(), "data: Array must have at least 1 items")
}

func TestErrorValidatingTestNode(t *testing.T) {
	c := readTestInput(t, "testdata/input-invalid-testnode.json")

	err := Match(c, s)
	assert.Contains(t, err.Error(), "test-node:node1-dsn.compute.internal TestNodeSample")
	assert.Contains(t, err.Error(), "capacity: capacity is required")
	assert.Contains(t, err.Error(), "test-node:node2-dsn.compute.internal TestNodeSample")
	assert.Contains(t, err.Error(), "cpuUsedCores: Invalid type. Expected: number, given: string")
}

func readTestInput(t *testing.T, filepath string) []byte {
	f, err := os.Open(filepath)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	c, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	return c
}
