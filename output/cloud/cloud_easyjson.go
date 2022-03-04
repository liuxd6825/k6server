// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package cloud

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	stats "go.k6.io/k6/stats"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson9def2ecdDecodeGoK6IoK6OutputCloud(in *jlexer.Lexer, out *samples) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		in.Skip()
		*out = nil
	} else {
		in.Delim('[')
		if *out == nil {
			if !in.IsDelim(']') {
				*out = make(samples, 0, 8)
			} else {
				*out = samples{}
			}
		} else {
			*out = (*out)[:0]
		}
		for !in.IsDelim(']') {
			var v1 *Sample
			if in.IsNull() {
				in.Skip()
				v1 = nil
			} else {
				if v1 == nil {
					v1 = new(Sample)
				}
				(*v1).UnmarshalEasyJSON(in)
			}
			*out = append(*out, v1)
			in.WantComma()
		}
		in.Delim(']')
	}
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9def2ecdEncodeGoK6IoK6OutputCloud(out *jwriter.Writer, in samples) {
	if in == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v2, v3 := range in {
			if v2 > 0 {
				out.RawByte(',')
			}
			if v3 == nil {
				out.RawString("null")
			} else {
				(*v3).MarshalEasyJSON(out)
			}
		}
		out.RawByte(']')
	}
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v samples) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9def2ecdEncodeGoK6IoK6OutputCloud(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *samples) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9def2ecdDecodeGoK6IoK6OutputCloud(l, v)
}
func easyjson9def2ecdDecodeGoK6IoK6OutputCloud1(in *jlexer.Lexer, out *SampleDataSingle) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "tags":
			if in.IsNull() {
				in.Skip()
				out.Tags = nil
			} else {
				if out.Tags == nil {
					out.Tags = new(stats.SampleTags)
				}
				if data := in.Raw(); in.Ok() {
					in.AddError((*out.Tags).UnmarshalJSON(data))
				}
			}
		case "time":
			out.Time = int64(in.Int64Str())
		case "type":
			if data := in.UnsafeBytes(); in.Ok() {
				in.AddError((out.Type).UnmarshalText(data))
			}
		case "value":
			out.Value = float64(in.Float64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9def2ecdEncodeGoK6IoK6OutputCloud1(out *jwriter.Writer, in SampleDataSingle) {
	out.RawByte('{')
	first := true
	_ = first
	if in.Tags != nil {
		const prefix string = ",\"tags\":"
		first = false
		out.RawString(prefix[1:])
		(*in.Tags).MarshalEasyJSON(out)
	}
	{
		const prefix string = ",\"time\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64Str(int64(in.Time))
	}
	{
		const prefix string = ",\"type\":"
		out.RawString(prefix)
		out.Raw((in.Type).MarshalJSON())
	}
	{
		const prefix string = ",\"value\":"
		out.RawString(prefix)
		out.Float64(float64(in.Value))
	}
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SampleDataSingle) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9def2ecdEncodeGoK6IoK6OutputCloud1(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SampleDataSingle) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9def2ecdDecodeGoK6IoK6OutputCloud1(l, v)
}
func easyjson9def2ecdDecodeGoK6IoK6OutputCloud2(in *jlexer.Lexer, out *SampleDataMap) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "tags":
			if in.IsNull() {
				in.Skip()
				out.Tags = nil
			} else {
				if out.Tags == nil {
					out.Tags = new(stats.SampleTags)
				}
				if data := in.Raw(); in.Ok() {
					in.AddError((*out.Tags).UnmarshalJSON(data))
				}
			}
		case "values":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.Values = make(map[string]float64)
				} else {
					out.Values = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v4 float64
					v4 = float64(in.Float64())
					(out.Values)[key] = v4
					in.WantComma()
				}
				in.Delim('}')
			}
		case "time":
			out.Time = int64(in.Int64Str())
		case "type":
			if data := in.UnsafeBytes(); in.Ok() {
				in.AddError((out.Type).UnmarshalText(data))
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9def2ecdEncodeGoK6IoK6OutputCloud2(out *jwriter.Writer, in SampleDataMap) {
	out.RawByte('{')
	first := true
	_ = first
	if in.Tags != nil {
		const prefix string = ",\"tags\":"
		first = false
		out.RawString(prefix[1:])
		(*in.Tags).MarshalEasyJSON(out)
	}
	if len(in.Values) != 0 {
		const prefix string = ",\"values\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		{
			out.RawByte('{')
			v5First := true
			for v5Name, v5Value := range in.Values {
				if v5First {
					v5First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v5Name))
				out.RawByte(':')
				out.Float64(float64(v5Value))
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"time\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64Str(int64(in.Time))
	}
	{
		const prefix string = ",\"type\":"
		out.RawString(prefix)
		out.Raw((in.Type).MarshalJSON())
	}
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SampleDataMap) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9def2ecdEncodeGoK6IoK6OutputCloud2(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SampleDataMap) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9def2ecdDecodeGoK6IoK6OutputCloud2(l, v)
}
func easyjson9def2ecdDecodeGoK6IoK6OutputCloud3(in *jlexer.Lexer, out *SampleDataAggregatedHTTPReqs) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "tags":
			if in.IsNull() {
				in.Skip()
				out.Tags = nil
			} else {
				if out.Tags == nil {
					out.Tags = new(stats.SampleTags)
				}
				if data := in.Raw(); in.Ok() {
					in.AddError((*out.Tags).UnmarshalJSON(data))
				}
			}
		case "type":
			out.Type = string(in.String())
		case "values":
			easyjson9def2ecdDecode(in, &out.Values)
		case "time":
			out.Time = int64(in.Int64Str())
		case "count":
			out.Count = uint64(in.Uint64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9def2ecdEncodeGoK6IoK6OutputCloud3(out *jwriter.Writer, in SampleDataAggregatedHTTPReqs) {
	out.RawByte('{')
	first := true
	_ = first
	if in.Tags != nil {
		const prefix string = ",\"tags\":"
		first = false
		out.RawString(prefix[1:])
		(*in.Tags).MarshalEasyJSON(out)
	}
	{
		const prefix string = ",\"type\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Type))
	}
	{
		const prefix string = ",\"values\":"
		out.RawString(prefix)
		easyjson9def2ecdEncode(out, in.Values)
	}
	{
		const prefix string = ",\"time\":"
		out.RawString(prefix)
		out.Int64Str(int64(in.Time))
	}
	{
		const prefix string = ",\"count\":"
		out.RawString(prefix)
		out.Uint64(uint64(in.Count))
	}
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SampleDataAggregatedHTTPReqs) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9def2ecdEncodeGoK6IoK6OutputCloud3(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SampleDataAggregatedHTTPReqs) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9def2ecdDecodeGoK6IoK6OutputCloud3(l, v)
}
func easyjson9def2ecdDecode(in *jlexer.Lexer, out *struct {
	Duration       AggregatedMetric `json:"http_req_duration"`
	Blocked        AggregatedMetric `json:"http_req_blocked"`
	Connecting     AggregatedMetric `json:"http_req_connecting"`
	TLSHandshaking AggregatedMetric `json:"http_req_tls_handshaking"`
	Sending        AggregatedMetric `json:"http_req_sending"`
	Waiting        AggregatedMetric `json:"http_req_waiting"`
	Receiving      AggregatedMetric `json:"http_req_receiving"`
	Failed         AggregatedRate   `json:"http_req_failed,omitempty"`
}) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "http_req_duration":
			easyjson9def2ecdDecodeGoK6IoK6OutputCloud4(in, &out.Duration)
		case "http_req_blocked":
			easyjson9def2ecdDecodeGoK6IoK6OutputCloud4(in, &out.Blocked)
		case "http_req_connecting":
			easyjson9def2ecdDecodeGoK6IoK6OutputCloud4(in, &out.Connecting)
		case "http_req_tls_handshaking":
			easyjson9def2ecdDecodeGoK6IoK6OutputCloud4(in, &out.TLSHandshaking)
		case "http_req_sending":
			easyjson9def2ecdDecodeGoK6IoK6OutputCloud4(in, &out.Sending)
		case "http_req_waiting":
			easyjson9def2ecdDecodeGoK6IoK6OutputCloud4(in, &out.Waiting)
		case "http_req_receiving":
			easyjson9def2ecdDecodeGoK6IoK6OutputCloud4(in, &out.Receiving)
		case "http_req_failed":
			easyjson9def2ecdDecodeGoK6IoK6OutputCloud5(in, &out.Failed)
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9def2ecdEncode(out *jwriter.Writer, in struct {
	Duration       AggregatedMetric `json:"http_req_duration"`
	Blocked        AggregatedMetric `json:"http_req_blocked"`
	Connecting     AggregatedMetric `json:"http_req_connecting"`
	TLSHandshaking AggregatedMetric `json:"http_req_tls_handshaking"`
	Sending        AggregatedMetric `json:"http_req_sending"`
	Waiting        AggregatedMetric `json:"http_req_waiting"`
	Receiving      AggregatedMetric `json:"http_req_receiving"`
	Failed         AggregatedRate   `json:"http_req_failed,omitempty"`
}) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"http_req_duration\":"
		out.RawString(prefix[1:])
		easyjson9def2ecdEncodeGoK6IoK6OutputCloud4(out, in.Duration)
	}
	{
		const prefix string = ",\"http_req_blocked\":"
		out.RawString(prefix)
		easyjson9def2ecdEncodeGoK6IoK6OutputCloud4(out, in.Blocked)
	}
	{
		const prefix string = ",\"http_req_connecting\":"
		out.RawString(prefix)
		easyjson9def2ecdEncodeGoK6IoK6OutputCloud4(out, in.Connecting)
	}
	{
		const prefix string = ",\"http_req_tls_handshaking\":"
		out.RawString(prefix)
		easyjson9def2ecdEncodeGoK6IoK6OutputCloud4(out, in.TLSHandshaking)
	}
	{
		const prefix string = ",\"http_req_sending\":"
		out.RawString(prefix)
		easyjson9def2ecdEncodeGoK6IoK6OutputCloud4(out, in.Sending)
	}
	{
		const prefix string = ",\"http_req_waiting\":"
		out.RawString(prefix)
		easyjson9def2ecdEncodeGoK6IoK6OutputCloud4(out, in.Waiting)
	}
	{
		const prefix string = ",\"http_req_receiving\":"
		out.RawString(prefix)
		easyjson9def2ecdEncodeGoK6IoK6OutputCloud4(out, in.Receiving)
	}
	if (in.Failed).IsDefined() {
		const prefix string = ",\"http_req_failed\":"
		out.RawString(prefix)
		easyjson9def2ecdEncodeGoK6IoK6OutputCloud5(out, in.Failed)
	}
	out.RawByte('}')
}
func easyjson9def2ecdDecodeGoK6IoK6OutputCloud5(in *jlexer.Lexer, out *AggregatedRate) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "count":
			out.Count = float64(in.Float64())
		case "nz_count":
			out.NzCount = float64(in.Float64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9def2ecdEncodeGoK6IoK6OutputCloud5(out *jwriter.Writer, in AggregatedRate) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"count\":"
		out.RawString(prefix[1:])
		out.Float64(float64(in.Count))
	}
	{
		const prefix string = ",\"nz_count\":"
		out.RawString(prefix)
		out.Float64(float64(in.NzCount))
	}
	out.RawByte('}')
}
func easyjson9def2ecdDecodeGoK6IoK6OutputCloud4(in *jlexer.Lexer, out *AggregatedMetric) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "min":
			out.Min = float64(in.Float64())
		case "max":
			out.Max = float64(in.Float64())
		case "avg":
			out.Avg = float64(in.Float64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9def2ecdEncodeGoK6IoK6OutputCloud4(out *jwriter.Writer, in AggregatedMetric) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"min\":"
		out.RawString(prefix[1:])
		out.Float64(float64(in.Min))
	}
	{
		const prefix string = ",\"max\":"
		out.RawString(prefix)
		out.Float64(float64(in.Max))
	}
	{
		const prefix string = ",\"avg\":"
		out.RawString(prefix)
		out.Float64(float64(in.Avg))
	}
	out.RawByte('}')
}
func easyjson9def2ecdDecodeGoK6IoK6OutputCloud6(in *jlexer.Lexer, out *Sample) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "data":
			if m, ok := out.Data.(easyjson.Unmarshaler); ok {
				m.UnmarshalEasyJSON(in)
			} else if m, ok := out.Data.(json.Unmarshaler); ok {
				_ = m.UnmarshalJSON(in.Raw())
			} else {
				out.Data = in.Interface()
			}
		case "type":
			out.Type = string(in.String())
		case "metric":
			out.Metric = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9def2ecdEncodeGoK6IoK6OutputCloud6(out *jwriter.Writer, in Sample) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"data\":"
		out.RawString(prefix[1:])
		if m, ok := in.Data.(easyjson.Marshaler); ok {
			m.MarshalEasyJSON(out)
		} else if m, ok := in.Data.(json.Marshaler); ok {
			out.Raw(m.MarshalJSON())
		} else {
			out.Raw(json.Marshal(in.Data))
		}
	}
	{
		const prefix string = ",\"type\":"
		out.RawString(prefix)
		out.String(string(in.Type))
	}
	{
		const prefix string = ",\"metric\":"
		out.RawString(prefix)
		out.String(string(in.Metric))
	}
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Sample) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9def2ecdEncodeGoK6IoK6OutputCloud6(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Sample) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9def2ecdDecodeGoK6IoK6OutputCloud6(l, v)
}
