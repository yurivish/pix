package pix

type posList struct {
	rest  []Pos
	first Pos
}

func (p *posList) insert(pos Pos) {
	p.rest = append(p.rest, pos)
}

func (p *posList) delete(pos Pos) bool {
	rest := p.rest
	n := len(rest)
	if p.first == pos {
		if n > 0 {
			// move a guy from rest to first
			p.first = rest[n-1]
			p.rest = rest[:n-1]
		} else {
			// the first position has been removed and there are no others.
			return true
		}
	} else {
		// remove the position from the rest list
		var found bool
		for i, x := range rest {
			if x == pos {
				rest[i] = rest[n-1]
				p.rest = rest[:n-1]
				found = true
				break
			}
		}
		if !found {
			panic("attempting to remove a non-existent position from the frontier")
		}
	}
	return false
}

func (p *posList) arbitrary() Pos {
	return p.first
}

/* for later, benchmark this first-less version:
package pix

type posList struct {
	rest []Pos
}

func (p *posList) insert(pos Pos) {
	p.rest = append(p.rest, pos)
}

func (p *posList) delete(pos Pos) bool {
	rest := p.rest
	n := len(rest)

	// remove the position from the rest list
	var found bool
	for i, x := range rest {
		if x == pos {
			rest[i] = rest[n-1]
			p.rest = rest[:n-1]
			found = true
			break
		}
	}
	if !found {
		panic("attempting to remove a non-existent position from the frontier")
	}

	return n == 1
}

func (p *posList) arbitrary() Pos {
	return p.rest[0]
}
*/
