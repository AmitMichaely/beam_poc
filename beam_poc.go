package main

import (
	"context"
	"fmt"
	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/io/textio"
	"github.com/apache/beam/sdks/go/pkg/beam/runners/direct"
	"github.com/apache/beam/sdks/go/pkg/beam/transforms/stats"
	"regexp"

	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/gcs"
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
)

var wordRE = regexp.MustCompile(`[a-zA-Z]+('[a-z])?`)

func init() {
	// beam.Init() is an initialization hook that must be called on startup.
	beam.Init()
}

func main() {
	// Create the Pipeline object and root scope.
	p := beam.NewPipeline()
	s := p.Root()

	// Apply the pipeline's transforms.

	// Concept #1: Invoke a root transform with the pipeline; in this case,
	// textio.Read to read a set of input text file. textio.Read returns a
	// PCollection where each element is one line from the input text
	// (one of of Shakespeare's texts).

	// This example reads a public data set consisting of the complete works
	// of Shakespeare.
	lines := textio.Read(s, "gs://apache-beam-samples/shakespeare/*")

	// Concept #2: Invoke a ParDo transform on our PCollection of text lines.
	// This ParDo invokes a DoFn (defined in-line) on each element that
	// tokenizes the text line into individual words. The ParDo returns a
	// PCollection of type string, where each element is an individual word in
	// Shakespeare's collected texts.
	words := beam.ParDo(s, func(line string, emit func(string)) {
		for _, word := range wordRE.FindAllString(line, -1) {
			emit(word)
		}
	}, lines)

	// Concept #3: Invoke the stats.Count transform on our PCollection of
	// individual words. The Count transform returns a new PCollection of
	// key/value pairs, where each key represents a unique word in the text.
	// The associated value is the occurrence count for that word.
	counted := stats.Count(s, words)

	// Use a ParDo to format our PCollection of word counts into a printable
	// string, suitable for writing to an output file. When each element
	// produces exactly one element, the DoFn can simply return it.
	formatted := beam.ParDo(s, func(w string, c int) string {
		return fmt.Sprintf("%s: %v", w, c)
	}, counted)

	// Concept #4: Invoke textio.Write at the end of the pipeline to write
	// the contents of a PCollection (in this case, our PCollection of
	// formatted strings) to a text file.
	textio.Write(s, "output/wordcounts.txt", formatted)

	// Run the pipeline on the direct runner.
	direct.Execute(context.Background(), p)
}
