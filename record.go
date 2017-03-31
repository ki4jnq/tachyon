package tachyon

type Record map[string]string

// Returns a list of of the Record's data sorted the same as f.Fields. It will
// return an interface{} type to correspond to Stmt.Exec's expectations.
func (r Record) orderedData(f *Fixture) []interface{} {
	list := make([]interface{}, 0, len(f.Fields))

	for _, fName := range f.Fields {
		if val, ok := r[fName]; ok {
			list = append(list, val)
		} else {
			list = append(list, "")
		}
	}
	return list
}
