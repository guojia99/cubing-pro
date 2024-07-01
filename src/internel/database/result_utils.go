package database

import (
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
)

func (c *convenient) sortByResults(
	allResults []result.Results,
	players []user.User,
) {

}

//(U2) R' F R F' (r U' R' U' R U' R') | (U') F (R U' R' U' R U R') U R U' R' F' |
//R U2 R' U' R U R' U2' R' F R F' | (R' U2' R U R' U' R U2) R f' U' f |
