using Go = import "go.capnp";
@0x85233c35877fd707; # Scehma ID
$Go.package("cpnp");
$Go.import("github.com/TheGreatSage/capntest/cpnp");


struct FakeMessage {
    email @0 :Text;
    ip @1 :Text;
    username @2 :Text;
    rfc3339 @3 :Text;
    unix @4 :Int64;
    uuid @5 :Text;
    ran @6 :UInt32;
}