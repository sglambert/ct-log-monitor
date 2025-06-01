
package models

type STH struct {
    TreeSize          uint64 `json:"tree_size"`
    PreviousTreeSize  uint64 `json:"previous_tree_size"`
    Timestamp         uint64 `json:"timestamp"`
    PreviousTimestamp uint64 `json:"previous_timestamp"`
}

type LogbookEntry struct {
    Name string `json:"name"`
    URL  string `json:"url"`
    STH  STH    `json:"sth"`
}

type CTGetSTHResponse struct {
    TreeSize          uint64 `json:"tree_size"`
    Timestamp         uint64 `json:"timestamp"`
    Sha256RootHash    string `json:"sha256_root_hash"`
    TreeHeadSignature string `json:"tree_head_signature"`
}

type LogEntry struct {
	LeafInput string `json:"leaf_input"`
	ExtraData string `json:"extra_data"`
}

type CTLogEntriesResponse struct {
	Entries []LogEntry `json:"entries"`
}

type MerkleTreeLeaf struct {
	Version        uint8
	LeafType       uint8
	TimestampedEntry TimestampedEntry
}

type TimestampedEntry struct {
	Timestamp  uint64
	EntryType  uint16
	ASN1Data   []byte
}