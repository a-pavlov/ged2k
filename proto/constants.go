package proto

const OP_EDONKEYHEADER byte = 0xE3
const OP_EDONKEYPROT byte = 0xE3
const OP_PACKEDPROT byte = 0xD4
const OP_EMULEPROT byte = 0xC5
const OP_KAD_COMPRESSED_UDP byte = 0xE5
const OP_KADEMLIAHEADER byte = 0xE4

const OP_LOGINREQUEST byte = 0x01  // <HASH 16><ID 4><PORT 2><1 Tag_set>
const OP_REJECT byte = 0x05        // (null)
const OP_GETSERVERLIST byte = 0x14 // (null)client->server
const OP_OFFERFILES byte = 0x15    // <count 4>(<HASH 16><ID 4><PORT 2><1
// Tag_set>)[count]
const OP_SEARCHREQUEST byte = 0x16 // <Query_Tree>
const OP_DISCONNECT byte = 0x18    // (not verified)
const OP_GETSOURCES byte = 0x19    // <HASH 16>
// v2 <HASH 16><SIZE_4> (17.3) (mandatory on 17.8)
// v2large <HASH 16><FILESIZE 4 byte 0)><FILESIZE 8>
// (17.9) (large files only)
const OP_SEARCH_USER byte = 0x1A     // <Query_Tree>
const OP_CALLBACKREQUEST byte = 0x1C // <ID 4>
// const OP_QUERY_CHATS = 0x1D, // (deprecated, not supported by server any
// longer)
// const OP_CHAT_MESSAGE = 0x1E, // (deprecated, not supported by server any
// longer)
// const OP_JOIN_ROOM = 0x1F, // (deprecated, not supported by server any
// longer)
const OP_QUERY_MORE_RESULT byte = 0x21 // ?
const OP_GETSOURCES_OBFU byte = 0x23
const OP_SERVERLIST byte = 0x32 // <count 1>(<IP 4><PORT
// 2>)[count]
// server->client
const OP_SEARCHRESULT byte = 0x33 // <count 4>(<HASH 16><ID 4><PORT 2><1
// Tag_set>)[count]
const OP_SERVERSTATUS byte = 0x34      // <USER 4><FILES 4>
const OP_CALLBACKREQUESTED byte = 0x35 // <IP 4><PORT 2>
const OP_CALLBACK_FAIL byte = 0x36     // (null notverified)
const OP_SERVERMESSAGE byte = 0x38     // <len 2><Message len>
// const OP_CHAT_ROOM_REQUEST = 0x39, // (deprecated, not supported by server
// any longer)
// const OP_CHAT_BROADCAST = 0x3A, // (deprecated, not supported by server any
// longer)
// const OP_CHAT_USER_JOIN = 0x3B, // (deprecated, not supported by server any
// longer)
// const OP_CHAT_USER_LEAVE = 0x3C, // (deprecated, not supported by server
// any longer)
// const OP_CHAT_USER = 0x3D, // (deprecated, not supported by server any
// longer)
const OP_IDCHANGE byte = 0x40     // <NEW_ID 4>
const OP_SERVERIDENT byte = 0x41  // <HASH 16><IP 4><PORT 2>{1 TAG_SET}
const OP_FOUNDSOURCES byte = 0x42 // <HASH 16><count 1>(<ID 4><PORT 2>)[count]
const OP_USERS_LIST byte = 0x43   // <count 4>(<HASH 16><ID 4><PORT 2><1
// Tag_set>)[count]
const OP_FOUNDSOURCES_OBFU byte = 0x44 // <HASH 16><count 1>(<ID 4><PORT 2><obf
// settings 1>(UserHash16 if
// obf&0x08))[count]

const ED2K_MAX_PACKET_SIZE uint32 = 125000

const HEADER_SIZE int = 6
