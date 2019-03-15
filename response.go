package splunksearch

import (
	"encoding/xml"
)

type MessagesResponse struct {
	Messages []Message `xml:"messages>msg"`
}

type Message struct {
	Type string `xml:"type,attr"`
	Msg  string `xml:",chardata"`
}

type Feed struct {
	TotalResults int     `xml:"totalResults"`
	ItemsPerPage int     `xml:"itemsPerPage"`
	StartIndex   int     `xml:"startIndex"`
	Entries      []Entry `xml:"entry"`
}

type Entry struct {
	Title   string `xml:"title"`
	Content SType  `xml:"content"`
}

type SType struct {
	Map  map[string]SType
	List []SType
	Str  string
}

type stype struct {
	Items []sitem `xml:"item"`
	Keys  []skey  `xml:"key"`
}

type sitem struct {
	List *stype `xml:"list"`
	Dict *stype `xml:"dict"`
	Str  string `xml:",chardata"`
}

type skey struct {
	Key string `xml:"name,attr"`
	sitem
}

func (s *sitem) toSType() (t SType) {
	switch {
	case s.List != nil:
		return s.List.toSType()
	case s.Dict != nil:
		return s.Dict.toSType()
	default:
		return SType{Str: s.Str}
	}
}

func (s *stype) toSType() (t SType) {
	switch {
	case s.Items != nil:
		l := make([]SType, len(s.Items))
		for i, v := range s.Items {
			l[i] = v.toSType()
		}
		t.List = l
	case s.Keys != nil:
		m := make(map[string]SType, len(s.Keys))
		for _, k := range s.Keys {
			m[k.Key] = k.toSType()
		}
		t.Map = m
	default:
		t.Str = ""
	}

	return
}

func (s *SType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var st sitem
	if err := d.DecodeElement(&st, &start); err != nil {
		return err
	}
	*s = st.toSType()

	return nil
}
