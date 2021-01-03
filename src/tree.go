package kagi

// count is saved at 0, and root right after it
func (db *DB_CONNECTION) loadDB() {
	db.getCount()
	db.root = db.getNodeAt(uint32(Int32Size) + 1)
}

func (db *DB_CONNECTION) createRootNode(k string, v string) {
	n := &Node{}

	n.isRoot = TRUE
	n.isDeleted = FALSE

	n.keys = make([]*Data, 1)
	n.keys[0] = &Data{data: []byte(k), size: int32(len(k))}
	n.numKeys = 1

	n.leaves = make([]*Leaf, 1)
	n.leaves[0] = NewLeaf(k, v)
	n.numLeaves = 1

	n.offset = uint32(Int32Size) + 1

	db.root = n
	db.count = 1
	db.writeNodeToFile(n)
}

func (db *DB_CONNECTION) insert(k string, v string) error {
	// if tree is empty add a new root node
	if db.count == 0 {
		db.createRootNode(k, v)
		return nil
	}

	newLeaf := NewLeaf(k, v)
	parent := db.searchNode(k, db.root)
	var keyFound string

	for i := 0; i < int(parent.numLeaves); i++ {
		// found leaf with correct key or no more leaves left
		if k <= string(parent.leaves[i].key.data) {
			keyFound = string(parent.leaves[i].key.data)
			break
		}
	}

	if k == keyFound {
		return KEY_ALREADY_EXISTS
	}

	parent.addLeaf(db, newLeaf)

	return nil
}

func (db *DB_CONNECTION) findLeaf(k string) (*Leaf, error) {
	parent := db.searchNode(k, db.root)
	index := 0

	for index = 0; index < int(parent.numLeaves); index++ {
		// found leaf with correct key or no more leaves left
		if string(parent.leaves[index].key.data) == k {
			break
		}
	}

	if string(parent.leaves[index].key.data) == k {
		return parent.leaves[index], nil
	}

	return nil, KEY_NOT_FOUND
}

// recursively traverse tree till we find node that has a leaves
func (db *DB_CONNECTION) searchNode(k string, currentNode *Node) *Node {
	if !currentNode.checkHasLeaf() {
		i := 0
		for i = 0; i < int(currentNode.numKeys); i++ {
			if k < string(currentNode.keys[i].data) {
				n := db.getNodeAt(currentNode.childOffsets[i])
				return db.searchNode(k, n)
			}
		}

		// search rightmost child
		n := db.getNodeAt(currentNode.childOffsets[i])
		return db.searchNode(k, n)
	}

	return currentNode
}

func (db *DB_CONNECTION) getNodeAt(offset uint32) *Node {
	b := make([]byte, BlockSize)
	db.readBytesAt(b, int64(offset))
	n := NodeFromBytes(b)

	return n
}

func (db *DB_CONNECTION) getCount() {
	b := make([]byte, Int32Size)
	db.readBytesAt(b, 0)
	db.count = Uint32FromBytes(b)

}

func (db *DB_CONNECTION) writeNodeToFile(n *Node) {
	nodeBytes := n.toBytes()
	db.writeBytesAt(nodeBytes, int64(n.offset))
}

// TODO delete node
//func (db *DB_CONNECTION) removeNode(k string) error {}

func (db *DB_CONNECTION) readBytesAt(b []byte, offset int64) {
	db.Lock()

	_, err1 := db.file.Seek(offset, 0)
	Check(err1)

	_, err2 := db.file.Read(b)
	Check(err2)

	db.Unlock()
}

func (db *DB_CONNECTION) writeBytesAt(b []byte, offset int64) {
	_, err1 := db.file.Seek(offset, 0)
	Check(err1)

	written, err2 := db.file.Write(b)
	Check(err2)

	if written < len(b) {
		Check(ERROR_WRITING)
	}
}
