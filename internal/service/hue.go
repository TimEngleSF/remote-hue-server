package service

import (
	"fmt"
	"slices"
	"strings"

	"github.com/amimof/huego"
)

type GroupNames []string

func (gn GroupNames) String() string {
	var formattedString string
	for _, name := range gn {
		formattedString += fmt.Sprintf("%s\n", name)
	}
	return formattedString
}

func (gn GroupNames) ArrayString() string {
	formattedString := "["
	for _, name := range gn {
		formattedString += fmt.Sprintf("'%s, '", name)
	}
	formattedString = strings.TrimSuffix(formattedString, ", ")
	formattedString += "]"
	return formattedString
}

type Group struct {
	huego.Group
}
type Groups []Group

func (gs *Groups) GroupStatusMessage(names GroupNames) string {
	msg := ""
	for _, g := range *gs {
		if slices.Contains(names, g.Name) {
			isOn := "Off"
			if g.State.On {
				isOn = "On"
				// Convert the brightness value to a percentage
				if g.State.Bri > 0 {
					brightnessPercentage := (float64(g.State.Bri) / 254.0) * 100
					msg += fmt.Sprintf("%v: %v, Brightness: %.0f%%\n", g.Name, isOn, brightnessPercentage)
				} else {
					msg += fmt.Sprintf("%v: %v\n", g.Name, isOn)
				}
			} else {
				msg += fmt.Sprintf("%v: %v\n", g.Name, isOn)
			}
		}
	}
	return msg
}

func (g Group) String() string {
	return fmt.Sprintf(
		`{"name": %s, "lights": %+v,"isOn": %v, "brightness": %d}`,
		g.Name, g.Lights, g.State.On, g.State.Bri,
	)
}

func (gs Groups) String() string {
	var sb strings.Builder
	for _, g := range gs {
		sb.WriteString(fmt.Sprintf("%v\n", g))
	}
	return sb.String()
}
