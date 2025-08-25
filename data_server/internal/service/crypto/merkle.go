package crypto

import "crypto/sha256"

// BuildMerkleRoot builds a binary Merkle root (dup last if odd).
func BuildMerkleRoot(leaves [][]byte) []byte {
  if len(leaves) == 0 { return nil }
  level := make([][]byte, len(leaves))
  copy(level, leaves)
  for len(level) > 1 {
    next := make([][]byte, 0, (len(level)+1)/2)
    for i := 0; i < len(level); i += 2 {
      if i+1 == len(level) {
        h := sha256.Sum256(append(level[i], level[i]...))
        next = append(next, h[:])
      } else {
        h := sha256.Sum256(append(level[i], level[i+1]...))
        next = append(next, h[:])
      }
    }
    level = next
  }
  return level[0]
}

type ProofStep struct {
  Sibling       []byte
  SiblingIsLeft bool
}

// BuildProof returns path for `index` against leaves (same tree rules).
func BuildProof(leaves [][]byte, index int) (leaf []byte, path []ProofStep, root []byte) {
  if len(leaves) == 0 || index < 0 || index >= len(leaves) { return nil, nil, nil }
  nodes := make([][]byte, len(leaves))
  copy(nodes, leaves)
  idx := index
  for len(nodes) > 1 {
    // collect sibling
    if idx%2 == 0 { // left
      sibIdx := idx + 1
      if sibIdx >= len(nodes) { // duplicate
        path = append(path, ProofStep{Sibling: nodes[idx], SiblingIsLeft: false})
      } else {
        path = append(path, ProofStep{Sibling: nodes[sibIdx], SiblingIsLeft: false})
      }
    } else {
      path = append(path, ProofStep{Sibling: nodes[idx-1], SiblingIsLeft: true})
    }

    // build next level
    next := make([][]byte, 0, (len(nodes)+1)/2)
    for i := 0; i < len(nodes); i += 2 {
      if i+1 == len(nodes) {
        h := sha256.Sum256(append(nodes[i], nodes[i]...))
        next = append(next, h[:])
      } else {
        h := sha256.Sum256(append(nodes[i], nodes[i+1]...))
        next = append(next, h[:])
      }
    }
    idx = idx / 2
    nodes = next
  }
  return leaves[index], path, nodes[0]
}
