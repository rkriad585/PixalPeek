package qrengine

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type pdfBuf struct {
	bytes.Buffer
}

func JPEGToPDF(jpeg []byte, dim int) ([]byte, error) {
	if len(jpeg) == 0 || dim <= 0 {
		return nil, fmt.Errorf("invalid jpeg payload for pdf")
	}

	var objs [][]byte
	add := func(b []byte) int {
		objs = append(objs, b)
		return len(objs)
	}

	catalogRef := add(nil)
	pagesRef := add(nil)
	pageRef := add(nil)
	imageRef := add(nil)
	contentRef := add(nil)

	objs[catalogRef-1] = fmt.Appendf(nil, "<< /Type /Catalog /Pages %d 0 R >>", pagesRef)
	objs[pagesRef-1] = fmt.Appendf(nil, "<< /Type /Pages /Kids [%d 0 R] /Count 1 >>", pageRef)
	objs[pageRef-1] = fmt.Appendf(nil,
		"<< /Type /Page /Parent %d 0 R /MediaBox [0 0 %d %d] /Resources << /XObject << /Im0 %d 0 R >> >> /Contents %d 0 R >>",
		pagesRef, dim, dim, imageRef, contentRef)
	objs[imageRef-1] = bytes.Join([][]byte{
		fmt.Appendf(nil, "<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length %d >>\nstream\n", dim, dim, len(jpeg)),
		jpeg,
		[]byte("\nendstream"),
	}, nil)

	content := fmt.Appendf(nil, "q\n%d 0 0 %d 0 0 cm\n/Im0 Do\nQ", dim, dim)
	objs[contentRef-1] = fmt.Appendf(nil, "<< /Length %d >>\nstream\n%s\nendstream", len(content), content)

	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n%\xE2\xE3\xCF\xD3\n")
	offsets := make([]int, len(objs))
	for i, obj := range objs {
		offsets[i] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xrefStart := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(objs)+1)
	for _, off := range offsets {
		fmt.Fprintf(&out, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root %d 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objs)+1, catalogRef, xrefStart)
	_ = binary.MaxVarintLen64
	return out.Bytes(), nil
}
