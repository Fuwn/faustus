package ui

import "github.com/charmbracelet/lipgloss"

var (
	Primary    = lipgloss.Color("#6B50FF")
	Secondary  = lipgloss.Color("#FF60FF")
	Tertiary   = lipgloss.Color("#68FFD6")
	Accent     = lipgloss.Color("#E8FE96")
	BgBase     = lipgloss.Color("#201F26")
	BgLighter  = lipgloss.Color("#2D2C35")
	BgSubtle   = lipgloss.Color("#3A3943")
	BgOverlay  = lipgloss.Color("#4D4C57")
	FgBase     = lipgloss.Color("#DFDBDD")
	FgMuted    = lipgloss.Color("#858392")
	FgHalfMute = lipgloss.Color("#BFBCC8")
	FgSubtle   = lipgloss.Color("#605F6B")
	FgBright   = lipgloss.Color("#F1EFEF")
	Success    = lipgloss.Color("#00FFB2")
	Error      = lipgloss.Color("#EB4268")
	Warning    = lipgloss.Color("#E8FE96")
	Info       = lipgloss.Color("#00A4FF")
	Blue       = lipgloss.Color("#00A4FF")
	Green      = lipgloss.Color("#00FFB2")
	GreenDark  = lipgloss.Color("#12C78F")
	Red        = lipgloss.Color("#FF577D")
	RedDark    = lipgloss.Color("#EB4268")
	Yellow     = lipgloss.Color("#E8FE96")
	Orange     = lipgloss.Color("#FF985A")
	Purple     = lipgloss.Color("#8B75FF")
	Cyan       = lipgloss.Color("#0ADCD9")
	Pink       = lipgloss.Color("#FF60FF")
	BaseStyle  = lipgloss.NewStyle().
			Foreground(FgBase)
	MutedStyle = lipgloss.NewStyle().
			Foreground(FgMuted)
	SubtleStyle = lipgloss.NewStyle().
			Foreground(FgSubtle)
	HeaderStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)
	LogoStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)
	TabStyle = lipgloss.NewStyle().
			Foreground(FgMuted).
			Padding(0, 2)
	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Padding(0, 2)
	ItemStyle = lipgloss.NewStyle().
			Foreground(FgBase).
			PaddingLeft(2)
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(FgBright).
				Background(BgSubtle).
				Bold(true).
				PaddingLeft(1).
				PaddingRight(1)
	CursorStyle = lipgloss.NewStyle().
			Foreground(Tertiary).
			Bold(true)
	TitleStyle = lipgloss.NewStyle().
			Foreground(FgBase).
			Bold(true)
	SummaryStyle = lipgloss.NewStyle().
			Foreground(FgHalfMute).
			Italic(true)
	MetaStyle = lipgloss.NewStyle().
			Foreground(FgSubtle)
	ProjectStyle = lipgloss.NewStyle().
			Foreground(Tertiary)
	TrashStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)
	ActiveStyle = lipgloss.NewStyle().
			Foreground(Success)
	SearchStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)
	SearchInputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(Primary).
				Padding(0, 1)
	HelpStyle = lipgloss.NewStyle().
			Foreground(FgSubtle)
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(FgMuted)
	ModalStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Background(BgLighter).
			Padding(1, 2)
	ConfirmStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(FgBase).
			Background(BgSubtle).
			Padding(0, 1)
	CountStyle = lipgloss.NewStyle().
			Foreground(Tertiary).
			Bold(true)
	PreviewStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(FgSubtle).
			Padding(0, 1)
	PreviewFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(Primary).
				Padding(0, 1)
	ListBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(FgSubtle).
			Padding(0, 1)
	ListBoxFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(Primary).
				Padding(0, 1)
	PreviewHeaderStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Bold(true)
	PreviewDividerStyle = lipgloss.NewStyle().
				Foreground(FgSubtle)
	UserRoleStyle = lipgloss.NewStyle().
			Foreground(Blue).
			Bold(true)
	UserContentStyle = lipgloss.NewStyle().
				Foreground(FgBase)
	AssistantRoleStyle = lipgloss.NewStyle().
				Foreground(GreenDark).
				Bold(true)
	AssistantContentStyle = lipgloss.NewStyle().
				Foreground(FgHalfMute)
	ToolRoleStyle = lipgloss.NewStyle().
			Foreground(Orange).
			Bold(true)
	ToolContentStyle = lipgloss.NewStyle().
				Foreground(FgMuted).
				Italic(true)
	ThinkingRoleStyle = lipgloss.NewStyle().
				Foreground(Purple).
				Bold(true)
	ThinkingContentStyle = lipgloss.NewStyle().
				Foreground(FgSubtle).
				Italic(true)
	HighlightStyle = lipgloss.NewStyle().
			Background(Accent).
			Foreground(BgBase).
			Bold(true)
	SearchResultStyle = lipgloss.NewStyle().
				Foreground(FgBase)
	SearchMatchStyle = lipgloss.NewStyle().
				Foreground(Accent).
				Bold(true)
	SearchContextStyle = lipgloss.NewStyle().
				Foreground(FgHalfMute)
)
