package snippet

import (
	"corgi/util"
	"encoding/json"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/kataras/iris/core/errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Snippet struct {
	Title   string      `json:"title"`
	Steps   []*StepInfo `json:"steps"`
	fileLoc string
}

type StepInfo struct {
	Command           string   `json:"command"`
	Description       string   `json:"description,omitempty"`
	ExecuteConcurrent bool     `json:"execute_concurrent"`
	TemplateFields    []string `json:"template_fields"`
}

type Answerable interface {
	AskQuestion(options ...interface{}) error
}

func scan(prompt string, defaultInp string) (string, error) {
	// create config
	config := &readline.Config{
		Prompt:            prompt,
		HistoryFile:       TempHistFile,
		HistorySearchFold: true,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
	}
	rl, err := readline.NewEx(config)
	if err != nil {
		return "", err
	}
	defer rl.Close()

	for {
		line, err := rl.ReadlineWithDefault(defaultInp)
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		return line, nil
	}
	return "", errors.New("cancelled")
}

// ################### Step related code ############################

func NewStepInfo(command string) *StepInfo {
	return &StepInfo{
		Command: command,
	}
}

func (step *StepInfo) AskQuestion(options ...interface{}) error {
	// set command
	cmd, err := scan(color.GreenString("Command: "), step.Command)
	if err != nil {
		return err
	}
	// TODO: read template from command
	step.Command = cmd
	// set description
	description, err := scan(color.GreenString("Description: "), "")
	if err != nil {
		return err
	}
	step.Description = description
	return nil
}

func (step *StepInfo) Execute() error {
	fmt.Printf("%s: %s\n", color.GreenString("Running"), color.YellowString(step.Command))
	commandsList := strings.Split(step.Command, "&&")
	for _, c := range commandsList {
		c = strings.TrimSpace(c)
		cmdName := strings.Split(c, " ")[0]
		cmdArgs := strings.Split(c, " ")[1:]
		cmd := exec.Command(cmdName, cmdArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			color.Red("[ Failed ]")
			return err
		}
	}
	color.Green("[ Success ]")
	return nil
}

// ################### Snippet related code ############################

func NewSnippet(title string, cmds []string) (*Snippet, error) {
	snippet := &Snippet{
		Title: title,
	}
	if err := snippet.AskQuestion(cmds); err != nil {
		return nil, err
	}
	return snippet, nil
}

func (snippet *Snippet) AskQuestion(options ...interface{}) error {
	// check options
	initialDefaultCmds := options[0].([]string)
	// ask about each step
	stepCount := 0
	steps := make([]*StepInfo, 0)
	for {
		color.Yellow("Step %d:", stepCount+1)
		var defaultCmd string
		if stepCount < len(initialDefaultCmds) {
			defaultCmd = initialDefaultCmds[stepCount]
		}
		step := NewStepInfo(defaultCmd)
		err := step.AskQuestion()
		if err != nil {
			return err
		}
		steps = append(steps, step)
		var addOneMoreStep bool
		for {
			addStepInp, err := scan(color.RedString("Add another step? (y/n): "), "")
			if err != nil {
				return err
			}
			if addStepInp == "y" {
				addOneMoreStep = true
			} else if addStepInp == "n" {
				addOneMoreStep = false
			} else {
				continue
			}
			break
		}
		fmt.Print("\n")
		if !addOneMoreStep {
			break
		}
		stepCount++
	}
	snippet.Steps = steps
	// ask about title if not set
	if snippet.Title == "" {
		title, err := scan(color.YellowString("Title: "), "")
		if err != nil {
			return err
		}
		snippet.Title = title
	}
	return nil
}

func (snippet *Snippet) Save(snippetsDir string) error {
	fmt.Printf("Saving snippet %s... ", snippet.Title)
	filePath := fmt.Sprintf("%s/%s.json", snippetsDir, strings.Replace(snippet.Title, " ", "_", -1))
	snippet.fileLoc = filePath
	data, err := json.Marshal(snippet)
	if err != nil {
		color.Red("Failed")
		return err
	}
	if err = ioutil.WriteFile(filePath, data, 0644); err != nil {
		color.Red("Failed")
		return err
	}
	color.Green("Success")
	return nil
}

func (snippet *Snippet) Execute() error {
	fmt.Println(color.GreenString("Start executing snippet \"%s\"...\n", snippet.Title))
	for idx, step := range snippet.Steps {
		fmt.Printf("%s: %s\n", color.GreenString("Step %d", idx+1), color.YellowString(step.Description))
		if err := step.Execute(); err != nil {
			return err
		}
		fmt.Println("")
	}
	return nil
}

func (snippet *Snippet) GetFilePath() string {
	return snippet.fileLoc
}

func LoadSnippet(filePath string) (*Snippet, error) {
	snippet := &Snippet{}
	if err := util.LoadJsonDataFromFile(filePath, snippet); err != nil {
		return nil, err
	}
	snippet.fileLoc = filePath
	return snippet, nil
}
