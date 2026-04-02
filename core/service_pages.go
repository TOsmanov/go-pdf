package core

import (
	"bytes"
	"fmt"

	"github.com/TOsmanov/go-pdf/internal/config"
)

func ToCPageHTML(params *Params, tocPage *string, cfg *config.Config) error {
	var b bytes.Buffer
	PrepareToC(params)
	err := cfg.ToCPageTemplate.Execute(&b, params.ToC)
	if err != nil {
		return err
	}

	*tocPage = b.String()
	return nil
}

func NewCounter() map[int]int {
	counter := map[int]int{
		2: 0,
		3: 0,
		4: 0,
		5: 0,
		6: 0,
	}
	return counter
}

func resetLvls(counter map[int]int, lvl int) map[int]int {
	for i := lvl + 1; i <= 6; i++ {
		counter[i] = 0
	}
	return counter
}

func sumPrevLvls(counter map[int]int, thisLvl int) int {
	sum := 0
	for i := thisLvl - 1; i >= 2; i-- {
		sum += counter[i]
	}
	return sum
}

func PrepareToC(params *Params) {
	for indP, p := range params.ToC {
		counter := NewCounter()
		symCount := NewCounter()
		prevLvl := 0
		var Indent int
		params.ToC[indP].Index = indP + 1
		for iH, h := range p.Headers {
			var lvls string
			if h.Lvl < prevLvl {
				counter = resetLvls(counter, h.Lvl)
			}

			counter[h.Lvl]++
			for i := 1; i <= h.Lvl; i++ {
				if counter[i] > 0 {
					lvls = fmt.Sprintf("%s.%d", lvls, counter[i])
				}
			}

			ind := fmt.Sprintf("%d%s", indP+1, lvls)
			symCount[h.Lvl] = len(ind) + 4 - h.Lvl

			Indent = sumPrevLvls(symCount, h.Lvl)

			params.ToC[indP].Headers[iH].Index = ind
			params.ToC[indP].Headers[iH].Class = fmt.Sprintf("lvl-%d", h.Lvl)
			params.ToC[indP].Headers[iH].Indent = Indent - 1
			prevLvl = params.ToC[indP].Headers[iH].Lvl
		}
	}
}

func TitlePageHTML(params *Params, titlePage *string, cfg *config.Config) error {
	var err error
	var b bytes.Buffer

	err = cfg.TitlePageTemplate.Execute(&b, params)
	if err != nil {
		return err
	}

	*titlePage = b.String()

	return nil
}
