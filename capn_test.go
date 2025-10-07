package main

import (
	"errors"
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
			buf, err := msg.MarshalThree(buf)
			if err != nil {
				b.Fatal(err)
			}
			if len(buf) == 0 {
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
