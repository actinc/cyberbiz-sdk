package gendocs

import "testing"

func TestStripHTML(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", "plain"},
		{"a<br>b<br/>c", "a\nb\nc"},
		{"訂單狀態<br> 「已開啟」: open<br> 「已結案」: closed", "訂單狀態\n「已開啟」: open\n「已結案」: closed"},
		{"<div class='x'><h3>Title</h3><ol><li>one</li><li>two</li></ol></div>", "Title\n- one\n- two"},
		{"a &amp; b &lt;c&gt;", "a & b <c>"},
		{"  spaced   out  ", "spaced out"},
		{"x\n\n\n\ny", "x\ny"},
	}
	for _, c := range cases {
		if got := stripHTML(c.in); got != c.want {
			t.Errorf("stripHTML(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
