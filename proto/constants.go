package proto

const OP_EDONKEYHEADER byte = 0xE3
const OP_EDONKEYPROT byte = 0xE3
const OP_PACKEDPROT byte = 0xD4
const OP_EMULEPROT byte = 0xC5
const OP_KAD_COMPRESSED_UDP byte = 0xE5
const OP_KADEMLIAHEADER byte = 0xE4

// server
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

const OP_HELLO byte = 0x01                // 0x10<HASH 16><ID 4><PORT 2><1 Tag_set>
const OP_SENDINGPART byte = 0x46          // <HASH 16><von 4><bis 4><Daten len:(von-bis)>
const OP_REQUESTPARTS byte = 0x47         // <HASH 16><von[3] 4*3><bis[3] 4*3>
const OP_FILEREQANSNOFIL byte = 0x48      // <HASH 16>
const OP_END_OF_DOWNLOAD byte = 0x49      // <HASH 16> // Unused for sending
const OP_ASKSHAREDFILES byte = 0x4A       // (null)
const OP_ASKSHAREDFILESANSWER byte = 0x4B // <count 4>(<HASH 16><ID 4><PORT 2><1 Tag_set>)[count]
const OP_HELLOANSWER byte = 0x4C          // <HASH 16><ID 4><PORT 2><1 Tag_set><SERVER_IP 4><SERVER_PORT 2>
const OP_CHANGE_CLIENT_ID byte = 0x4D     // <ID_old 4><ID_new 4> // Unused for sending
const OP_MESSAGE byte = 0x4E              // <len 2><Message len>
const OP_SETREQFILEID byte = 0x4F         // <HASH 16>
const OP_FILESTATUS byte = 0x50           // <HASH 16><count 2><status(bit array) len:((count+7)/8)>
const OP_HASHSETREQUEST byte = 0x51       // <HASH 16>
const OP_HASHSETANSWER byte = 0x52        // <count 2><HASH[count] 16*count>
const OP_STARTUPLOADREQ byte = 0x54       // <HASH 16>
const OP_ACCEPTUPLOADREQ byte = 0x55      // (null)
const OP_CANCELTRANSFER byte = 0x56       // (null)
const OP_OUTOFPARTREQS byte = 0x57        // (null)
const OP_REQUESTFILENAME byte = 0x58      // <HASH 16>    (more correctly file_name_request)
const OP_REQFILENAMEANSWER byte = 0x59    // <HASH 16><len 4><NAME len>
const OP_CHANGE_SLOT byte = 0x5B          // <HASH 16> // Not used for sending
const OP_QUEUERANK byte = 0x5C            // <wert  4> (slot index of the request) // Not used for sending
const OP_ASKSHAREDDIRS byte = 0x5D        // (null)
const OP_ASKSHAREDFILESDIR byte = 0x5E    // <len 2><Directory len>
const OP_ASKSHAREDDIRSANS byte = 0x5F     // <count 4>(<len 2><Directory len>)[count]
const OP_ASKSHAREDFILESDIRANS byte = 0x60 // <len 2><Directory len><count 4>(<HASH 16><ID 4><PORT 2><1 T
const OP_ASKSHAREDDENIEDANS byte = 0x61   // (null)

const OP_EMULEINFO byte = 0x01       //
const OP_EMULEINFOANSWER byte = 0x02 //
const OP_COMPRESSEDPART byte = 0x40  //
const OP_QUEUERANKING byte = 0x60    // <RANG 2>
const OP_FILEDESC byte = 0x61        // <len 2><NAME len>
const OP_VERIFYUPSREQ byte = 0x71    // (never used)
const OP_VERIFYUPSANSWER byte = 0x72 // (never used)
const OP_UDPVERIFYUPREQ byte = 0x73  // (never used)
const OP_UDPVERIFYUPA byte = 0x74    // (never used)
const OP_REQUESTSOURCES byte = 0x81  // <HASH 16>
const OP_ANSWERSOURCES byte = 0x82   //
const OP_REQUESTSOURCES2 byte = 0x83 // <HASH 16>
const OP_ANSWERSOURCES2 byte = 0x84  //
const OP_PUBLICKEY byte = 0x85       // <len 1><pubkey len>
const OP_SIGNATURE byte = 0x86       // v1: <len 1><signature len>
// v2:<len 1><signature len><sigIPused 1>
const OP_SECIDENTSTATE byte = 0x87  // <state 1><rndchallenge 4>
const OP_REQUESTPREVIEW byte = 0x90 // <HASH 16> // Never used for sending on aMule
const OP_PREVIEWANSWER byte = 0x91  // <HASH 16><frames 1>{frames * <len 4><frame len>} // Never used for sending on aMule
const OP_MULTIPACKET byte = 0x92
const OP_MULTIPACKETANSWER byte = 0x93

// OP_PEERCACHE_QUERY        byte = 0x94 // Unused on aMule - no PeerCache
// OP_PEERCACHE_ANSWER       byte = 0x95 // Unused on aMule - no PeerCache
// OP_PEERCACHE_ACK          byte = 0x96 // Unused on aMule - no PeerCache
const OP_PUBLICIP_REQ byte = 0x97
const OP_PUBLICIP_ANSWER byte = 0x98
const OP_CALLBACK byte = 0x99 // <HASH 16><HASH 16><uint 16>
const OP_REASKCALLBACKTCP byte = 0x9A
const OP_AICHREQUEST byte = 0x9B // <HASH 16><uint16><HASH aichhashlen>
const OP_AICHANSWER byte = 0x9C  // <HASH 16><uint16><HASH aichhashlen> <data>
const OP_AICHFILEHASHANS byte = 0x9D
const OP_AICHFILEHASHREQ byte = 0x9E
const OP_BUDDYPING byte = 0x9F
const OP_BUDDYPONG byte = 0xA0
const OP_COMPRESSEDPART_I64 byte = 0xA1 // <HASH 16><von 8><size 4><Data len:size>
const OP_SENDINGPART_I64 byte = 0xA2    // <HASH 16><start 8><end 8><Data len:(end-start)>
const OP_REQUESTPARTS_I64 byte = 0xA3   // <HASH 16><start[3] 8*3><end[3] 8*3>
const OP_MULTIPACKET_EXT byte = 0xA4
const OP_CHATCAPTCHAREQ byte = 0xA5
const OP_CHATCAPTCHARES byte = 0xA6

const ED2K_MAX_PACKET_SIZE int = 125000

const HEADER_SIZE int = 6

const SO_EMULE int = 0
const SO_CDONKEY int = 1
const SO_LXMULE int = 2
const SO_AMULE int = 3
const SO_SHAREAZA int = 4
const SO_EMULEPLUS int = 5
const SO_HYDRANODE int = 6
const SO_NEW2_MLDONKEY int = 0x0a
const SO_LPHANT int = 0x14
const SO_NEW2_SHAREAZA int = 0x28
const SO_EDONKEYHYBRID int = 0x32
const SO_EDONKEY int = 0x33
const SO_MLDONKEY int = 0x34
const SO_OLDEMULE int = 0x35
const SO_UNKNOWN int = 0x36
const SO_NEW_SHAREAZA int = 0x44
const SO_NEW_MLDONKEY int = 0x98
const SO_LIBED2K int = 0x99
const SO_QMULE int = 0xA0
const SO_COMPAT_UNK int = 0xFF

const PIECE_SIZE int = 9728000
const PIECE_SIZE_UINT64 uint64 = 9728000
const BLOCK_SIZE int = 190 * 1024                    // 190kb = PIECE_SIZE/50
const BLOCK_SIZE_UINT64 uint64 = 190 * 1024          // 190kb = PIECE_SIZE/50
const BLOCKS_PER_PIECE int = PIECE_SIZE / BLOCK_SIZE // 50
const HIGHEST_LOWID_ED2K uint32 = 16777216
const REQUEST_QUEUE_SIZE int = 3
