package notification

import (
	"fmt"
	"strings"
	"time"

	"fibo-monitor/config"
	"fibo-monitor/indicator"
	"fibo-monitor/signal"
)

type MessageCard struct {
	Config config.MessageCardConfig
}

func NewMessageCard(cfg config.MessageCardConfig) *MessageCard {
	return &MessageCard{
		Config: cfg,
	}
}

// Lark Card Structure
type LarkCard struct {
	MsgType string   `json:"msg_type"`
	Card    CardBody `json:"card"`
}

type CardBody struct {
	Header   CardHeader   `json:"header"`
	Elements []interface{} `json:"elements"`
}

type CardHeader struct {
	Title    TagText `json:"title"`
	Template string  `json:"template"` // blue, red, etc.
}

type TagText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type DivElement struct {
	Tag    string        `json:"tag"`
	Text   TagText       `json:"text"`
	Fields []FieldObject `json:"fields,omitempty"`
}

type FieldObject struct {
	IsShort bool    `json:"is_short"`
	Text    TagText `json:"text"`
}

type ActionElement struct {
	Tag     string         `json:"tag"`
	Actions []ButtonObject `json:"actions"`
}

type ButtonObject struct {
	Tag   string  `json:"tag"`
	Text  TagText `json:"text"`
	Url   string  `json:"url"`
	Type  string  `json:"type"` // default, primary, danger
	Value map[string]interface{} `json:"value,omitempty"`
}

func (m *MessageCard) BuildLarkMessage(sig signal.Signal) LarkCard {
	// Theme color mapping: 
	// Golden Cross (Bullish) -> Blue/Green -> "blue" or "turquoise"
	// Death Cross (Bearish) -> Red -> "red" or "carmine"
	template := "blue"
	titleText := "📈 金叉信号 (做多)"
	if sig.Type == indicator.DeathCross {
		template = "red"
		titleText = "📉 死叉信号 (做空)"
	}

	// Content Fields
	fields := []FieldObject{
		{
			IsShort: true,
			Text: TagText{
				Tag:     "lark_md",
				Content: fmt.Sprintf("**交易对**\n%s", sig.Symbol),
			},
		},
		{
			IsShort: true,
			Text: TagText{
				Tag:     "lark_md",
				Content: fmt.Sprintf("**周期**\n%s", sig.Interval),
			},
		},
		{
			IsShort: true,
			Text: TagText{
				Tag:     "lark_md",
				Content: fmt.Sprintf("**当前价格**\n%.2f", sig.Price),
			},
		},
	}

	if m.Config.IncludeEmaValues {
		fields = append(fields, FieldObject{
			IsShort: true,
			Text: TagText{
				Tag:     "lark_md",
				Content: fmt.Sprintf("**EMA Short**\n%.2f", sig.ShortEMA),
			},
		})
		fields = append(fields, FieldObject{
			IsShort: true,
			Text: TagText{
				Tag:     "lark_md",
				Content: fmt.Sprintf("**EMA Long**\n%.2f", sig.LongEMA),
			},
		})
	}

	if m.Config.IncludeTimestamp {
		fields = append(fields, FieldObject{
			IsShort: false, // Timestamp usually long
			Text: TagText{
				Tag:     "lark_md",
				Content: fmt.Sprintf("**时间**\n%s", sig.Timestamp.Format("2006-01-02 15:04:05")),
			},
		})
	}
    
    // Buttons
    var actions []ButtonObject
    for _, btn := range m.Config.LarkSpecific.Buttons {
        // Replace placeholders in URL
        url := strings.ReplaceAll(btn.URL, "{symbol}", sig.Symbol)
        
        actions = append(actions, ButtonObject{
            Tag: "button",
            Text: TagText{
                Tag: "plain_text",
                Content: btn.Text,
            },
            Url: url,
            Type: "primary",
        })
    }
    
    elements := []interface{}{
        DivElement{
            Tag: "div",
            Fields: fields,
        },
        DivElement{
            Tag: "hr",
        },
        ActionElement{
            Tag: "action",
            Actions: actions,
        },
    }

	return LarkCard{
		MsgType: "interactive",
		Card: CardBody{
			Header: CardHeader{
				Template: template,
				Title: TagText{
					Tag:     "plain_text",
					Content: titleText,
				},
			},
			Elements: elements,
		},
	}
}
