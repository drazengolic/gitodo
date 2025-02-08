/*
Copyright © 2025 Drazen Golic

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// The UI spaghetti goes here
package ui

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/drazengolic/gitodo/base"
	"github.com/drazengolic/gitodo/shell"
	"github.com/gen2brain/beeep"
)

// state modes
const (
	ModeTodoItems int = iota
	ModeQueue
	ModeInput
)

var (
	// control channels
	resetChan chan struct{}
	exitChan  chan struct{}

	// various styles
	greenText = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#007700", Dark: "#00ff00"})

	checkMark = greenText.SetString("X").Bold(true).String()

	boldText = lipgloss.NewStyle().Bold(true)

	commitedBox = greenText.SetString("commited").String()

	dimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#777777"))

	redText = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))

	orangeText = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#ff7500", Dark: "#ffa500"})

	timerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#ff0000")).
			Foreground(lipgloss.Color("#ffffff"))

	timerStoppedStyle = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#777777"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#777777", Dark: "#000000"})
)

func init() {
	resetChan = make(chan struct{})
	exitChan = make(chan struct{})
}

// operations that need confirmation via prompt

type opDelTodoItem int
type opDelQueueItem int
type opPushStash int
type opPopStash int

type TickMsg time.Time

// model is a structure that represents the state of the entire ui app
type model struct {
	todoItems    []todoItem
	queueItems   []todoItem
	cursor       int
	ready        bool
	viewport     viewport.Model
	mode         int
	proj         base.Project
	queueProjId  int
	errorMsg     string
	env          *shell.DirEnv
	db           *base.TodoDb
	prompt       string
	pendingOp    any
	showHelp     bool
	screenWidth  int
	screenHeight int
	showTodoId   bool
	timeTotal    int
	timerActive  bool
}

// initialModel creates the initial model from the data and the environment
func initialModel(env *shell.DirEnv, db *base.TodoDb) model {
	todoProjId := db.FetchProjectId(env.ProjDir, env.Branch)
	queueProjId := db.FetchProjectId(env.ProjDir, "*")
	todoItems := []todoItem{}
	queueItems := []todoItem{}

	stash, err := shell.GetStashItems()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	proj := db.GetProject(todoProjId)

	err = db.TodoItems(todoProjId, func(t base.Todo) {
		todoItems = append(todoItems, todoItem{
			id:       t.Id,
			task:     t.Task,
			done:     t.DoneAt.Valid,
			commited: t.CommitedAt.Valid,
			stash:    stash[t.Id],
		})
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = db.TodoItems(queueProjId, func(t base.Todo) {
		queueItems = append(queueItems, todoItem{
			id:       t.Id,
			task:     t.Task,
			done:     t.DoneAt.Valid,
			commited: t.CommitedAt.Valid,
			stash:    stash[t.Id],
		})
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	timeTotal, err := db.GetProjectTime(todoProjId)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	cursor := 0
	for i, t := range todoItems {
		if !t.done {
			cursor = i
			break
		}
	}

	model := model{
		todoItems:   todoItems,
		queueItems:  queueItems,
		mode:        ModeTodoItems,
		cursor:      cursor,
		proj:        proj,
		queueProjId: queueProjId,
		env:         env,
		db:          db,
		showHelp:    false,
		timeTotal:   timeTotal,
	}

	te := db.GetLatestTimeEntry()
	if te != nil && te.ProjectId == proj.Id && te.Action == base.TimesheetActionStart {
		model.timerActive = true
	}

	return model
}

func (m model) Init() tea.Cmd {
	if m.timerActive {
		return doTick()
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case TickMsg:
		m.timeTotal++
		return m, doTick()

	case tea.KeyMsg:

		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q", "Q":
			go func() { exitChan <- struct{}{} }()
			return m, nil

		// moving up
		case "up", "k", "K":
			if m.cursor > 0 {
				m.cursor--
			} else if m.cursor == 0 && m.mode == ModeQueue && len(m.todoItems) > 0 {
				m.stateMode(ModeTodoItems)
				m.cursor = len(m.todoItems) - 1
			} else if m.cursor == 0 && m.mode == ModeTodoItems && len(m.queueItems) > 0 {
				m.cursor = len(m.queueItems) - 1
				m.stateMode(ModeQueue)
				m.viewport.GotoBottom()
			}

		// moving down
		case "down", "j", "J":
			if m.mode == ModeTodoItems && m.cursor < len(m.todoItems)-1 {
				m.cursor++
			} else if m.mode == ModeQueue && m.cursor < len(m.queueItems)-1 {
				m.cursor++
			} else if m.mode == ModeTodoItems && m.cursor == len(m.todoItems)-1 && len(m.queueItems) > 0 {
				m.cursor = 0
				m.stateMode(ModeQueue)
			} else if m.mode == ModeQueue && m.cursor == len(m.queueItems)-1 && len(m.todoItems) > 0 {
				m.cursor = 0
				m.stateMode(ModeTodoItems)
				m.viewport.GotoTop()
			}

		// toggle "done"
		case "enter", " ":
			if m.mode == ModeTodoItems {
				done := !m.todoItems[m.cursor].done
				err := m.db.TodoDone(m.todoItems[m.cursor].id, done)
				if err != nil {
					m.errorMsg = err.Error()
				} else {
					m.todoItems[m.cursor].done = done
				}
			}

		// move item to the top of the list
		case "t", "T":
			if m.mode == ModeTodoItems {
				item := m.todoItems[m.cursor]
				err := m.db.ChangePosition(item.id, m.cursor+1, 1)
				if err != nil {
					m.errorMsg = err.Error()
				} else {
					m.todoItems = slices.Insert(slices.Delete(m.todoItems, m.cursor, m.cursor+1), 0, item)
					m.cursor = 0
				}
			}
		// shift item to the one step above
		case "ctrl+up", "ctrl+k":
			if m.mode == ModeTodoItems && m.cursor > 0 {
				item := m.todoItems[m.cursor]
				err := m.db.ChangePosition(item.id, m.cursor+1, m.cursor)
				if err != nil {
					m.errorMsg = err.Error()
				} else {
					swapItem := m.todoItems[m.cursor-1]
					m.todoItems[m.cursor-1] = item
					m.todoItems[m.cursor] = swapItem
					m.cursor--
				}
			}
		// shift item to the one step below
		case "ctrl+down", "ctrl+j":
			if m.mode == ModeTodoItems && m.cursor < len(m.todoItems)-1 {
				item := m.todoItems[m.cursor]
				err := m.db.ChangePosition(item.id, m.cursor+1, m.cursor+2)
				if err != nil {
					m.errorMsg = err.Error()
				} else {
					swapItem := m.todoItems[m.cursor+1]
					m.todoItems[m.cursor+1] = item
					m.todoItems[m.cursor] = swapItem
					m.cursor++
				}
			}
		// delete item
		case "d", "D":
			if m.mode == ModeTodoItems && len(m.todoItems) > 0 {
				if m.todoItems[m.cursor].done {
					beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
					break
				}
				m.stateMode(ModeInput)
				m.prompt = "delete todo item: are you sure? (y/n) "
				m.pendingOp = opDelTodoItem(m.cursor)
			} else if m.mode == ModeQueue && len(m.queueItems) > 0 {
				m.stateMode(ModeInput)
				m.prompt = "delete queue item: are you sure? (y/n) "
				m.pendingOp = opDelQueueItem(m.cursor)
			}

		// confirm prompt and do the pending op
		case "y", "Y":
			if m.mode != ModeInput {
				break
			}

			switch m.pendingOp.(type) {
			case opDelTodoItem:
				index := int(m.pendingOp.(opDelTodoItem))
				item := m.todoItems[index]
				err := m.db.Delete(item.id)
				if err == nil {
					m.todoItems = slices.Delete(m.todoItems, index, index+1)
				} else {
					m.errorMsg = err.Error()
				}
				m.mode = ModeTodoItems
				if len(m.todoItems) == 0 {
					m.cursor = 0
					m.stateMode(ModeQueue)
				}
			case opDelQueueItem:
				index := int(m.pendingOp.(opDelQueueItem))
				item := m.todoItems[index]
				err := m.db.Delete(item.id)
				if err == nil {
					m.queueItems = slices.Delete(m.queueItems, index, index+1)
				} else {
					m.errorMsg = err.Error()
				}
				m.mode = ModeQueue
				if len(m.queueItems) == 0 {
					m.cursor = 0
					m.stateMode(ModeTodoItems)
				}
			case opPushStash:
				index := int(m.pendingOp.(opPushStash))
				item := m.todoItems[index]
				m.stateMode(ModeTodoItems)
				err := shell.PushStash(item.id)
				if err != nil {
					m.errorMsg = err.Error()
					break
				}
				stashes, err := shell.GetStashItems()
				if err != nil {
					m.errorMsg = err.Error()
					break
				}
				m.todoItems[index].stash = stashes[item.id]

			case opPopStash:
				index := int(m.pendingOp.(opPopStash))
				item := m.todoItems[index]
				m.stateMode(ModeTodoItems)
				err := shell.PopStash(item.stash)
				if err != nil {
					m.errorMsg = err.Error()
					break
				}
				m.todoItems[index].stash = ""
			}
		// cancel prompt, clear pending op
		case "n", "N":
			if m.mode != ModeInput {
				break
			}

			switch m.pendingOp.(type) {
			case opDelQueueItem:
				m.stateMode(ModeQueue)
			default:
				m.stateMode(ModeTodoItems)
			}
		// edit item in the external editor
		case "e", "E":
			var coll []todoItem
			if m.mode == ModeTodoItems {
				coll = m.todoItems
			} else if m.mode == ModeQueue {
				coll = m.queueItems
			} else {
				break
			}
			if len(coll) == 0 {
				break
			}

			tmp, err := shell.NewTmpFileString(coll[m.cursor].task)
			if err != nil {
				m.errorMsg = err.Error()
				break
			}

			defer tmp.Delete()
			defer func() { go func() { resetChan <- struct{}{} }() }()

			err = tmp.Edit(m.env.Editor, 0)
			if err != nil {
				m.errorMsg = err.Error()
				break
			}
			txt := strings.TrimSpace(tmp.ReadAll())
			if txt != "" {
				err = m.db.UpdateTask(coll[m.cursor].id, txt)
			}

			if err != nil {
				m.errorMsg = err.Error()
			} else {
				coll[m.cursor].task = txt
			}

		// move to/from queue
		case "m", "M":
			if m.mode == ModeTodoItems && len(m.todoItems) > 0 {
				item := m.todoItems[m.cursor]
				if item.done {
					beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
					break
				}
				err := m.db.MoveTodo(item.id, m.queueProjId)

				if err == nil {
					m.queueItems = append(m.queueItems, item)
					m.todoItems = slices.Delete(m.todoItems, m.cursor, m.cursor+1)
					m.cursor--
					if m.cursor < 0 {
						m.cursor = 0
					}
					if len(m.todoItems) == 0 {
						m.stateMode(ModeQueue)
					}
				} else {
					m.errorMsg = err.Error()
				}

			} else if m.mode == ModeQueue && len(m.queueItems) > 0 {
				item := m.queueItems[m.cursor]
				err := m.db.MoveTodo(item.id, m.proj.Id)

				if err == nil {
					m.todoItems = append(m.todoItems, item)
					m.queueItems = slices.Delete(m.queueItems, m.cursor, m.cursor+1)
					m.cursor--
					if m.cursor < 0 {
						m.cursor = 0
					}
					if len(m.queueItems) == 0 {
						m.stateMode(ModeTodoItems)
					}
				} else {
					m.errorMsg = err.Error()
				}
			}
		// toggle help display
		case "?", "h", "H":
			m.showHelp = !m.showHelp
			m.updateHeight()
		// save stash
		case "s", "S":
			if m.mode == ModeTodoItems {
				item := m.todoItems[m.cursor]
				if item.stash != "" {
					beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
					break
				}
				m.prompt = "push changes to stash? (y/n) "
				m.stateMode(ModeInput)
				m.pendingOp = opPushStash(m.cursor)
			}
		// pop stash
		case "p", "P":
			item := m.todoItems[m.cursor]
			if m.mode == ModeTodoItems && item.stash != "" {
				m.prompt = "pop changes from stash? (y/n) "
				m.stateMode(ModeInput)
				m.pendingOp = opPopStash(m.cursor)
			}
		// render todo item ids for advanced purposes
		case "#":
			m.showTodoId = !m.showTodoId
		}

	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.screenHeight = msg.Height
		headerHeight := m.getHeaderHeight()
		footerHeight := m.getFooterHeight()
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = false
			m.viewport.KeyMap.PageDown.Unbind()
			m.viewport.KeyMap.PageUp.Unbind()
			m.ready = true
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

	}

	m.viewport.SetContent(m.Content())

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m model) Content() string {
	builder := &strings.Builder{}

	// to-do items section

	builder.WriteString(boldText.Render("  TO-DO LIST:"))
	builder.WriteString("\n\n")

	itemWidth := m.viewport.Width - 6

	for i, choice := range m.todoItems {
		selected := m.mode == ModeTodoItems && m.cursor == i

		cursor := " "
		if selected {
			cursor = boldText.Render(">")
		}

		var checked string
		switch {
		case selected && choice.done:
			checked = boldText.Render("[") + checkMark + boldText.Render("]")
		case selected:
			checked = boldText.Render("[ ]")
		case choice.done:
			checked = "[" + checkMark + "]"
		default:
			checked = "[ ]"
		}

		// Render the row
		builder.WriteString(cursor)
		builder.WriteRune(' ')
		builder.WriteString(checked)
		builder.WriteRune(' ')
		builder.WriteString(choice.Render(selected, m.showTodoId, itemWidth, "      "))
		builder.WriteRune('\n')

	}

	if len(m.queueItems) == 0 {
		return builder.String()
	}

	// queued items section
	itemWidth = m.viewport.Width - 2

	builder.WriteString("\n")
	builder.WriteString(boldText.Render("  QUEUE:"))
	builder.WriteString("\n\n")

	for i, choice := range m.queueItems {
		selected := m.mode == ModeQueue && m.cursor == i
		cursor := " " // no cursor
		if selected {
			cursor = boldText.Render(">")
		}

		// Render the row
		builder.WriteString(cursor)
		builder.WriteRune(' ')
		builder.WriteString(choice.Render(selected, false, itemWidth, "  "))
		builder.WriteRune('\n')
	}

	return builder.String()
}

func (m model) headerView() string {
	if !m.ready {
		return ""
	}

	style := lipgloss.NewStyle().
		Width(m.viewport.Width).
		Padding(0, 2).
		AlignHorizontal(lipgloss.Center).Bold(true)

	b := strings.Builder{}

	if m.proj.Name != "" && m.proj.Name != m.proj.Branch {
		b.WriteString(style.Render(m.proj.Name))
		b.WriteRune('\n')
	}

	linew := m.viewport.Width - lipgloss.Width(m.env.Branch) - 2
	b.WriteString(dimmedStyle.Render(strings.Repeat("─", linew/2)))
	b.WriteRune(' ')
	b.WriteString(m.env.Branch)
	b.WriteRune(' ')
	b.WriteString(dimmedStyle.Render(strings.Repeat("─", linew/2+linew%2)))
	b.WriteRune('\n')

	if m.timeTotal > 0 || m.timerActive {
		timew := m.viewport.Width - 8
		secs := base.FormatSeconds(m.timeTotal)
		b.WriteString(strings.Repeat(" ", timew/2))
		if m.timerActive {
			b.WriteString(timerStyle.Render(secs))
		} else {
			b.WriteString(timerStoppedStyle.Render(secs))
		}

		b.WriteString(strings.Repeat(" ", timew/2+timew%2))
	}

	return b.String()
}

// footerView renders messages and a help table when enabled
func (m model) footerView() string {
	if !m.ready {
		return ""
	}

	var style = lipgloss.NewStyle().
		Width(m.viewport.Width).
		Padding(0, 0)

	b := strings.Builder{}

	if m.showHelp && m.mode != ModeInput {
		// help text per mode
		var keyMap [][]string

		if m.mode == ModeTodoItems {
			keyMap = [][]string{
				{"Quit", "Q"},
				{"Up", "K"},
				{"Down", "J"},
				{"Toggle done", "⎵"},
				{"Push up", "^K"},
				{"Push down", "^J"},
				{"Push to top", "T"},
				{"Edit", "E"},
				{"Delete", "D"},
				{"Move to queue", "M"},
				{"Stash", "S"},
				{"Pop stash", "P"},
			}
		} else {
			keyMap = [][]string{
				{"Quit", "Q"},
				{"Up", "K"},
				{"Down", "J"},
				{"Make todo", "M"},
				{"Edit", "E"},
				{"Delete", "D"},
			}
		}

		keys := make([]string, 0, len(keyMap))
		for _, k := range keyMap {
			if len(k[1]) == 2 {
				keys = append(keys, k[1]+" "+dimmedStyle.Render(k[0]))
			} else {
				keys = append(keys, " "+k[1]+" "+dimmedStyle.Render(k[0]))
			}
		}

		rows := slices.Collect(slices.Chunk(keys, 4))

		t := table.New().
			Width(m.viewport.Width).
			Height(len(rows) + 2).
			Border(lipgloss.HiddenBorder()).
			Rows(rows...)

		b.WriteString(dimmedStyle.Render(strings.Repeat("─", m.viewport.Width)))
		b.WriteString(t.String())
		b.WriteRune('\n')
	}

	// render errors or prompts first if any
	switch {
	case m.errorMsg != "":
		b.WriteString(style.Render(dimmedStyle.Render(strings.Repeat("─", m.viewport.Width)) + "\n  " + redText.Render(m.errorMsg)))
	case m.mode == ModeInput:
		b.WriteString(style.Render(dimmedStyle.Render(strings.Repeat("─", m.viewport.Width)) + "\n  " + orangeText.Render(m.prompt)))
	case m.mode == ModeTodoItems:
		b.WriteString(dimmedStyle.Render(strings.Repeat("─", m.viewport.Width)))
		b.WriteString(dimmedStyle.Render("\n  to-do items: toggle help with 'h' or '?'"))
	case m.mode == ModeQueue:
		b.WriteString(dimmedStyle.Render(strings.Repeat("─", m.viewport.Width)))
		b.WriteString(dimmedStyle.Render("\n  queue items: toggle help with 'h' or '?'"))
	default:
		b.WriteString(dimmedStyle.Render(strings.Repeat("─", m.viewport.Width)))
		b.WriteRune('\n')
	}
	return b.String()
}

// getHeaderHeight calculates header height from the state
// because lipgloss.Height is not reliable
func (m model) getHeaderHeight() int {
	var h int
	if m.proj.Name == "" || m.proj.Name == m.proj.Branch {
		h = 2
	}

	h = (len([]rune(m.proj.Name))-2)/m.screenWidth + 3

	if m.timeTotal > 0 || m.timerActive {
		h += 1
	}

	return h
}

// getHeaderHeight calculates footer height from the state
// because lipgloss.Height is not reliable
func (m model) getFooterHeight() int {
	var h int
	switch {
	case m.showHelp && m.mode == ModeTodoItems:
		h = 7
	case m.showHelp && m.mode == ModeQueue:
		h = 6
	default:
		h = 2
	}

	if m.timeTotal > 0 || m.timerActive {
		h -= 1
	}

	return h
}

// updateHeight updates the height of the viewport
func (m *model) updateHeight() {
	if m.ready {
		m.viewport.Height = m.screenHeight - m.getHeaderHeight() - m.getFooterHeight()
	}
}

// stateMode changes the state mode and applies side effects to the model
func (m *model) stateMode(mode int) {
	m.mode = mode
	if mode != ModeInput {
		m.prompt = ""
		m.pendingOp = nil
	}
	m.errorMsg = ""
	m.updateHeight()
}

func doTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// RunTodoListUI creates and runs the bubbletea program
func RunTodoListUI(env *shell.DirEnv, db *base.TodoDb) {

	var p *tea.Program
	model := initialModel(env, db)
	shouldRun := true

	// restart or quit via channels
	// because executing i.e. vim to edit stuff will mess up the alt screen
	// and the only solution is to restart the ui program
	go func() {
		for {
			select {
			case <-resetChan:
				p.Quit()
			case <-exitChan:
				shouldRun = false
				p.Quit()
				break
			}
		}
	}()

	for shouldRun {
		p = tea.NewProgram(
			model,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		if _, err := p.Run(); err != nil {
			fmt.Printf("There's been an error: %v", err)
			os.Exit(1)
		}
	}
}
