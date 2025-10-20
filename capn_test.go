package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	// "os"
	"testing"

	"capnproto.org/go/capnp/v3"
	bebop "wellquite.org/bebop/runtime"

	"google.golang.org/protobuf/proto"

	"github.com/TheGreatSage/capntest/capnpext"
	"github.com/TheGreatSage/capntest/cpnp"
	"github.com/TheGreatSage/capntest/prob"

	"github.com/TheGreatSage/capntest/beop"
)

func TestMarshal(t *testing.T) {
	msg := newMessage(t)
	data, err := msg.Cpn.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	m2, err := capnp.Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}
	fk, err := cpnp.ReadRootFakeMessage(m2)
	if err != nil {
		t.Fatal(err)
	}
	em, err := fk.Email()
	if err != nil {
		t.Fatal(err)
	}
	e2 := msg.Beo.Email
	if em != e2 {
		t.Fatal("email does not match")
	}
}

func BenchmarkNewMessage(b *testing.B) {
	buf, _ := capnp.NewSingleSegmentMessage(nil)
	fk := NewFakeSage()
	b.Run("No Write", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			seg, err := buf.Reset(buf.Arena)
			if err != nil {
				b.Fatal(err)
			}
			_, err = cpnp.NewRootFakeMessage(seg)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Write", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			seg, err := buf.Reset(buf.Arena)
			if err != nil {
				b.Fatal(err)
			}
			fake, err := cpnp.NewRootFakeMessage(seg)
			if err != nil {
				b.Fatal(err)
			}
			e1 := fake.SetEmail(fk.Email)
			e2 := fake.SetIp(fk.IP)
			e3 := fake.SetUsername(fk.Name)
			e4 := fake.SetRfc3339(fk.RFC)
			fake.SetUnix(fk.Unix)
			e5 := fake.SetUuid(fk.UUID)
			fake.SetRan(fk.Rand)
			if errors.Join(e1, e2, e3, e4, e5) != nil {
				b.Fail()
			}
		}
	})
}

type Messages struct {
	Fake *FakeSage
	Cpn  *capnp.Message
	Beo  beop.FakeMessage
	Pro  *prob.FakeMessage
}

func newMessage(tb testing.TB) *Messages {
	cmsg, seg := capnp.NewSingleSegmentMessage(nil)
	fk := NewFakeSage()

	bep := beop.FakeMessage{}
	fake, err := cpnp.NewRootFakeMessage(seg)
	if err != nil {
		tb.Fatal(err)
	}
	err1 := fake.SetEmail(fk.Email)
	err2 := fake.SetIp(fk.IP)
	err3 := fake.SetUsername(fk.Name)
	err4 := fake.SetRfc3339(fk.RFC)
	fake.SetUnix(fk.Unix)
	err5 := fake.SetUuid(fk.UUID)
	if errors.Join(err1, err2, err3, err4, err5) != nil {
		tb.Fatal(err)
	}
	_, err = fake.Email()
	if err != nil {
		tb.Fatal(err)
	}

	fake.SetRan(fk.Rand)
	bep.Email = fk.Email
	bep.Ip = fk.IP
	bep.Username = fk.Name
	bep.Rfc3339 = fk.RFC
	bep.Unix = fk.Unix
	bep.Uuid = fk.UUID
	bep.Ran = fk.Rand

	pro := prob.FakeMessage{
		Email:    fk.Email,
		Ip:       fk.IP,
		Username: fk.Name,
		Rfc3339:  fk.RFC,
		Unix:     fk.Unix,
		Uuid:     fk.UUID,
		Ran:      fk.Rand,
	}

	var msg Messages
	msg.Fake = fk
	msg.Cpn = cmsg
	msg.Beo = bep
	msg.Pro = &pro

	return &msg
}

func BenchmarkUnmarshal(b *testing.B) {
	msg := newMessage(b)

	// log.Printf("cpn %v\n", msg.Cpn.Ran())
	buf, err := msg.Cpn.Marshal()
	// var buf bytes.Buffer
	// err := capnp.NewEncoder(&buf).Encode(msg.Cpn)
	if err != nil {
		b.Fatal(err)
	}
	readmsg, _ := capnp.NewSingleSegmentMessage(nil)
	readbuf := make([]byte, 8*1024)

	reuse := NewCReuse()

	var bepbuf []byte
	bepbuf, err = msg.Beo.MarshalBebop(bepbuf)
	if err != nil {
		b.Fatal(err)
	}

	probuf, err := proto.Marshal(msg.Pro)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Unmarshal", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			m2, err := capnp.Unmarshal(buf)
			if err != nil {
				b.Fatal(err)
			}
			d, err := cpnp.ReadRootFakeMessage(m2)
			deserializeCheck(b, &d, msg.Fake, err)
		}
	})
	b.Run("UnmarshalZeroTo", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			err = capnp.UnmarshalZeroTo(readmsg, &readbuf, buf)
			if err != nil {
				b.Fatal(err)
			}
			d, err := cpnp.ReadRootFakeMessage(readmsg)
			deserializeCheck(b, &d, msg.Fake, err)
		}
	})
	b.Run("UnmarshalZeroThree", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			err = capnp.UnmarshalZeroThree(readmsg, buf)
			if err != nil {
				b.Fatal(err)
			}
			d, err := cpnp.ReadRootFakeMessage(readmsg)
			deserializeCheck(b, &d, msg.Fake, err)
		}
	})
	b.Run("Deserialize", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			dat, err := Deserialize(reuse, buf, cpnp.ReadRootFakeMessage)
			deserializeCheck(b, &dat, msg.Fake, err)
		}
	})
	b.Run("DeserializeZeroThree", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			dat, err := DeserializeZero(reuse, buf, cpnp.ReadRootFakeMessage)
			deserializeCheck(b, &dat, msg.Fake, err)
		}
	})
	b.Run("DeserializeZeroTo", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			dat, err := DeserializeZeroTo(reuse, buf, cpnp.ReadRootFakeMessage)
			deserializeCheck(b, &dat, msg.Fake, err)
		}
	})
	b.Run("Beop", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var bepp beop.FakeMessage
			_, err = bepp.UnmarshalBebop(bepbuf)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("DesccBeop", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var bepp beop.FakeMessage
			_, err := Desc(&bepp, bepbuf)
			if err != nil {
				b.Fatal(err)
			}
			if bepp.Email != msg.Fake.Email {
				b.Fail()
			}
		}
	})
	b.Run("Pro", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var ppro prob.FakeMessage
			err := proto.Unmarshal(probuf, &ppro)
			if err != nil {
				b.Fatal(err)
			}
			if ppro.Email != msg.Fake.Email {
				b.Fail()
			}
		}
	})
}

func Desc[T bebop.Bebop](bp T, data []byte) (int, error) {
	return bp.UnmarshalBebop(data)
}

func deserializeCheck(tb testing.TB, cpn *cpnp.FakeMessage, fks *FakeSage, err error) {
	if err != nil {
		tb.Fatal(err)
	}
	if cpn.Ran() != fks.Rand {
		log.Printf("Bad Rand: %d vs %d", cpn.Ran(), fks.Rand)
		tb.FailNow()
	}
	if !cpn.IsValid() {
		log.Println("Bad IsVlaid")
		tb.FailNow()
	}
}

func TestCapnpMarshals(t *testing.T) {
	msg := newMessage(t)
	reuse := NewCReuse()
	t.Run("Capnp Marshal", func(t *testing.T) {
		data, err := msg.Cpn.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		checkAllCapnpDeserialize(t, reuse, msg.Fake, data)
	})
	// t.Run("Capnp MarshalTo", func(tt *testing.T) {
	// 	tt.Skip("This function doesn't actually work it seems.")
	// 	buf := make([]byte, 8*1024)
	// 	buf, err := msg.Cpn.MarshalTo(buf)
	// 	if err != nil {
	// 		tt.Log("MarshalTo Error")
	// 		tt.Fatal(err)
	// 	}
	// 	checkAllCapnpDeserialize(tt, reuse, msg.Fake, buf)
	// })
	t.Run("Capnp MarshalThree", func(t *testing.T) {
		buf := make([]byte, 8*1024)
		_, err := msg.Cpn.MarshalThree(buf)
		if err != nil {
			t.Fatal(err)
		}
		checkAllCapnpDeserialize(t, reuse, msg.Fake, buf)
	})
	t.Run("Capnpext MarshalTo", func(t *testing.T) {
		buf := make([]byte, 8*1024)
		_, err := capnpext.MarshalTo(msg.Cpn, buf)
		if err != nil {
			t.Fatal(err)
		}
		checkAllCapnpDeserialize(t, reuse, msg.Fake, buf)
	})
	t.Run("Capnpext MarshalThree", func(t *testing.T) {
		buf := make([]byte, 8*1024)
		_, err := capnpext.MarshalThree(msg.Cpn, buf)
		if err != nil {
			t.Fatal(err)
		}
		checkAllCapnpDeserialize(t, reuse, msg.Fake, buf)
	})
}

func checkAllCapnpDeserialize(t *testing.T, reuse *CapnpReuse, check *FakeSage, data []byte) {
	d1, err := Deserialize(reuse, data, cpnp.ReadRootFakeMessage)
	deserializeCheck(t, &d1, check, err)
	d2, err := DeserializeZero(reuse, data, cpnp.ReadRootFakeMessage)
	deserializeCheck(t, &d2, check, err)
	d3, err := DeserializeZeroTo(reuse, data, cpnp.ReadRootFakeMessage)
	deserializeCheck(t, &d3, check, err)
}

func BenchmarkMarshal(b *testing.B) {
	msg := newMessage(b)
	fk := msg.Cpn
	bep := msg.Beo
	pro := msg.Pro
	buf := make([]byte, 8*1024)
	b.Run("Marshal", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			msg := fk
			data, err := msg.Marshal()
			if err != nil {
				b.Fatal(err)
			}
			if len(data) == 0 {
				b.Fail()
			}
		}
	})
	b.Run("MarshalTo", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			msg := fk
			buf, err := msg.MarshalTo(buf)
			if err != nil {
				b.Fatal(err)
			}
			if len(buf) == 0 {
				b.Fail()
			}
		}
	})
	b.Run("MarshalThree", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			msg := fk
			m, err := msg.MarshalThree(buf)
			if err != nil {
				b.Fatal(err)
			}
			if m == 0 {
				b.Fail()
			}
		}
	})
	b.Run("NewMarshalTo", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			msg := fk
			i, err := capnpext.MarshalTo(msg, buf)
			if err != nil {
				b.Fatal(err)
			}
			if i == 0 {
				b.Fail()
			}
		}
	})
	b.Run("NewMarshalThree", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			msg := fk
			i, err := capnpext.MarshalThree(msg, buf)
			if err != nil {
				b.Fatal(err)
			}
			if i == 0 {
				b.Fail()
			}
		}
	})
	b.Run("Beop", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			buf, err := bep.MarshalBebop(buf)
			if err != nil {
				b.Fatal(err)
			}
			if len(buf) == 0 {
				b.Fail()
			}
		}
	})
	b.Run("Pro", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			buf, err := proto.Marshal(pro)
			if err != nil {
				b.Fatal(err)
			}
			if len(buf) == 0 {
				b.Fail()
			}
		}
	})
}

// BenchmarkMarshalThreeStream
// A full test using MarshalThree and UnmarshalThree.
//
// Writes to io.Writer then Reads and deserializes the message.
func BenchmarkMarshalThreeStream(b *testing.B) {
	msg := newMessage(b)
	fk := msg.Cpn
	buffer := make([]byte, 8*1024)
	stream := new(bytes.Buffer)
	stream.Grow(8 * 1024)

	msgBuf, _ := capnp.NewSingleSegmentMessage(nil)

	readBuffer := make([]byte, 8*1024)

	type packetHeader struct {
		Id  uint16
		Len uint32
	}

	const packetHeaderLength = 6

	var header [packetHeaderLength]byte
	var pheader packetHeader

	b.ReportAllocs()
	for b.Loop() {
		// Not resetting stream fails test.
		stream.Reset()

		buf := buffer[:cap(buffer)]
		payload := buf[packetHeaderLength:]
		n, err := fk.MarshalThree(payload)
		if err != nil {
			b.Fatal(err)
		}

		// Pretend this is an actual stream

		binary.LittleEndian.PutUint16(buf[0:2], uint16(0xFF))
		binary.LittleEndian.PutUint32(buf[2:packetHeaderLength], uint32(n))
		_, err = stream.Write(buf[:n+packetHeaderLength])
		if err != nil {
			b.Fatal(err)
		}

		// Have to read the header first
		// binary.Read actually doers an alloc so have to do this.
		_, err = io.ReadFull(stream, header[:])
		if err != nil {
			b.Fatal(err)
		}
		pheader.Id = binary.LittleEndian.Uint16(header[:2])
		pheader.Len = binary.LittleEndian.Uint32(header[2:6])

		if pheader.Id != uint16(0xFF) {
			b.Fatal("Wrong packet ID", pheader.Id)
		}

		n, err = stream.Read(readBuffer)
		if err != nil {
			b.Fatal(err)
		}

		err = capnp.UnmarshalZeroThree(msgBuf, readBuffer[:n])
		if err != nil {
			b.Fatal(err)
		}

		fake, err := cpnp.ReadRootFakeMessage(msgBuf)
		if err != nil {
			b.Fatal(err)
		}

		if !fake.IsValid() {
			b.Fatal("message not valid")
		}

		if fake.Ran() != msg.Fake.Rand {
			b.Fatal("fake message not valid")
		}
	}
}
