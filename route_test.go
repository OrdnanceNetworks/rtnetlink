package rtnetlink

import (
	"bytes"
	"net"
	"reflect"
	"testing"

	"golang.org/x/sys/unix"
)

// Tests will only pass on little endian machines

func TestRouteMessageMarshalBinary(t *testing.T) {
	tests := []struct {
		name string
		m    Message
		b    []byte
		err  error
	}{
		{
			name: "empty",
			m:    &RouteMessage{},
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "no attributes",
			m: &RouteMessage{
				Family:    unix.AF_INET,
				DstLength: 8,
				Type:      unix.RTN_UNICAST,
			},
			b: []byte{
				0x02, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "attributes",
			m: &RouteMessage{
				Attributes: RouteAttributes{
					Dst:      net.ParseIP("10.0.0.0"),
					Gateway:  net.ParseIP("10.10.10.10"),
					OutIface: 4,
				},
			},
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x01, 0x00,
				0x0a, 0x00, 0x00, 0x00, 0x08, 0x00, 0x05, 0x00,
				0x0a, 0x0a, 0x0a, 0x0a, 0x08, 0x00, 0x04, 0x00,
				0x04, 0x00, 0x00, 0x00,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := tt.m.MarshalBinary()

			if want, got := tt.err, err; want != got {
				t.Fatalf("unexpected error:\n- want: %v\n-  got: %v", want, got)
			}
			if err != nil {
				return
			}

			if want, got := tt.b, b; !bytes.Equal(want, got) {
				t.Fatalf("unexpected Message bytes:\n- want: [%# x]\n-  got: [%# x]", want, got)
			}
		})
	}
}

func TestRouteMessageUnmarshalBinary(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
		m    Message
		err  error
	}{
		{
			name: "empty",
			err:  errInvalidRouteMessage,
		},
		{
			name: "short",
			b:    make([]byte, 3),
			err:  errInvalidRouteMessage,
		},
		{
			name: "invalid attr",
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x04, 0x00, 0x01, 0x00, 0x04, 0x00, 0x02, 0x00,
				0x05, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			err: errInvalidRouteMessageAttr,
		},
		{
			name: "data",
			b: []byte{
				0x02, 0x08, 0x00, 0x00, 0xfe, 0x04, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x01, 0x00,
				0x0a, 0x00, 0x00, 0x00, 0x08, 0x00, 0x07, 0x00,
				0x0a, 0x64, 0x0a, 0x01, 0x08, 0x00, 0x04, 0x00,
				0x05, 0x00, 0x00, 0x00,
			},
			m: &RouteMessage{
				Family:    2,
				DstLength: 8,
				Table:     unix.RT_TABLE_MAIN,
				Protocol:  unix.RTPROT_STATIC,
				Scope:     unix.RT_SCOPE_UNIVERSE,
				Type:      unix.RTN_UNICAST,
				Attributes: RouteAttributes{
					Dst:      net.IP{0x0a, 0x00, 0x00, 0x00},
					Src:      net.IP{0x0a, 0x64, 0x0a, 0x01},
					OutIface: 5,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RouteMessage{}
			err := (m).UnmarshalBinary(tt.b)

			if want, got := tt.err, err; want != got {
				t.Fatalf("unexpected error:\n- want: %v\n-  got: %v", want, got)
			}
			if err != nil {
				return
			}

			if want, got := tt.m, m; !reflect.DeepEqual(want, got) {
				t.Fatalf("unexpected Message:\n- want: %#v\n-  got: %#v", want, got)
			}
		})
	}
}
