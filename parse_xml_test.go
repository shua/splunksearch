package splunksearch

import (
	//"fmt"
	"encoding/xml"
	"io/ioutil"
	"log"
	"testing"
)

func Test_SType(t *testing.T) {

	blob := `
	<content>
	<s:dict>
		<s:key name="a">1</s:key>
		<s:key name="b">2</s:key>
		<s:key name="c">
			<s:list>
				<s:item>dog</s:item>
				<s:item>cat</s:item>
			</s:list>
		</s:key>
	</s:dict>
	</content>
	`

	var s SType

	if err := xml.Unmarshal([]byte(blob), &s); err != nil {
		t.Error(err)
	}

	if s.Map == nil {
		t.Errorf("Failed to parse top-level map")
	}
	if s.Map["a"].Str != "1" {
		t.Errorf("failed to parse a")
	}
	if s.Map["b"].Str != "2" {
		t.Error("failed to parse b")
	}
	if s.Map["c"].List == nil {
		t.Errorf("failed to parse c")
	}
	c := s.Map["c"].List
	if len(c) != 2 {
		t.Errorf("found len of %d, expected %d", len(c), 2)
	}
}

func Test_Feed(t *testing.T) {
	testFeed := `
	<feed>
		<entry>
			<content>
				<s:dict>
				</s:dict>
			</content>
		</entry>
	</feed>
	`

	var x SplunkFeed

	if err := xml.Unmarshal([]byte(testFeed), &x); err != nil {
		log.Fatal(err)
	}
}

func Test_SplunkResponse(t *testing.T) {
	response := `
	<response>
		<messages>
			<msg type="ERROR">something went wrong</msg>
		</messages>
	</response>
	`

	var x SplunkResponse

	if err := xml.Unmarshal([]byte(response), &x); err != nil {
		log.Fatal(err)
	}

	if x.Response == nil || len(x.Response) == 0 {
		t.Errorf("failed to parse response")
	}

	if x.Response[0].Type != "ERROR" || x.Response[0].Msg != "something went wrong" {
		t.Errorf("failed to parse msg")
	}
}
