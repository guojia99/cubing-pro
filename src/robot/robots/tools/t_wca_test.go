package tools

import (
	"fmt"
	"testing"

	"github.com/guojia99/cubing-pro/src/robot/types"
)

func TestTWca_Do(t *testing.T) {
	t.Run("person", func(t *testing.T) {
		tt := &TWca{}

		out, err := tt.Do(types.InMessage{
			Message: "Wca 徐永浩",
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, msg := range out.Message {
			fmt.Println(msg)
		}
	})

	t.Run("pk1", func(t *testing.T) {
		tt := &TWca{}

		out, err := tt.Do(types.InMessage{
			Message: "wca-pk 徐永浩-徐梓翼",
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, msg := range out.Message {
			fmt.Println(msg)
		}
	})

	t.Run("pk2", func(t *testing.T) {
		tt := &TWca{}

		out, err := tt.Do(types.InMessage{
			Message: "wca-pk 徐梓翅膀-徐永浩",
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, msg := range out.Message {
			fmt.Println(msg)
		}
	})

	t.Run("pk3", func(t *testing.T) {
		tt := &TWca{}

		out, err := tt.Do(types.InMessage{
			Message: "wca-pk 郭泽嘉-陈樑",
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, msg := range out.Message {
			fmt.Println(msg)
		}
	})

	t.Run("pk4", func(t *testing.T) {
		tt := &TWca{}

		out, err := tt.Do(types.InMessage{
			Message: "wca-pk 陈樑-陈樑",
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, msg := range out.Message {
			fmt.Println(msg)
		}
	})

	t.Run("cx1", func(t *testing.T) {
		tt := &TWca{}
		out, err := tt.Do(types.InMessage{
			Message: "wcacx 徐永浩-郭泽嘉",
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, msg := range out.Message {
			fmt.Println(msg)
		}
	})

	t.Run("cx2", func(t *testing.T) {
		tt := &TWca{}
		out, err := tt.Do(types.InMessage{
			Message: "wcacx 郭泽嘉-徐永浩",
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, msg := range out.Message {
			fmt.Println(msg)
		}
	})
}
