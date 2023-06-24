package parser

// TODO: use those functions during parsing

func zeroOrOne(p *parser, callback func() bool) {
	p.setCheckpoint()
	if callback() {
		p.commit()
	} else {
		p.rollback()
	}
}

func zeroOrMore(p *parser, callback func() bool) {
	for {
		p.setCheckpoint()
		if callback() {
			p.commit()
		} else {
			p.rollback()
			break
		}
	}
}

func oneOrMore(p *parser, callback func() bool) bool {
	p.setCheckpoint()
	if callback() {
		p.commit()
	} else {
		p.rollback()
		return false
	}

	zeroOrMore(p, callback)
	return true
}

func oneOf(p *parser, callbacks ...func() bool) bool {
	for _, callback := range callbacks {
		p.setCheckpoint()
		if callback() {
			p.commit()
			return true
		}
		p.rollback()
	}
	return false
}

func sequence(p *parser, callbacks ...func() bool) bool {
	for _, callback := range callbacks {
		p.setCheckpoint()
		if !callback() {
			p.rollback()
			return false
		}
		p.commit()
	}
	return true
}
