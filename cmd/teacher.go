package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"canvas-cli/internal/ui"
)

// runGrade grades (and/or comments on) a student's submission.
//
// Usage:
//
//	canvas-cli grade <course_id> <assignment_id> <student_id> <score> [--comment "text"]
//	canvas-cli grade <course_id> <assignment_id> <student_id> --comment "text"   (comment only)
//
// The <score> may be a point value ("18"), a percentage ("95%"), a letter
// grade ("A-"), or "complete"/"incomplete" depending on the assignment's
// grading type — Canvas interprets submission[posted_grade] accordingly.
func runGrade(args []string) {
	if len(args) < 3 {
		ui.Error("usage: canvas-cli grade <course_id> <assignment_id> <student_id> <score> [--comment \"text\"]")
		os.Exit(1)
	}

	courseID := args[0]
	assignID := args[1]
	studentID := args[2]
	remaining := args[3:]

	var score, comment string
	var hasScore bool

	for i := 0; i < len(remaining); i++ {
		switch remaining[i] {
		case "--comment":
			if i+1 < len(remaining) {
				comment = remaining[i+1]
				i++
			} else {
				ui.Error("--comment requires text")
				os.Exit(1)
			}
		default:
			// First non-flag argument is treated as the score.
			if !hasScore {
				score = remaining[i]
				hasScore = true
			}
		}
	}

	if !hasScore && comment == "" {
		ui.Error("provide a <score> and/or a --comment \"text\"")
		os.Exit(1)
	}

	form := url.Values{}
	if hasScore {
		form.Set("submission[posted_grade]", score)
	}
	if comment != "" {
		form.Set("comment[text_comment]", comment)
	}

	endpoint := fmt.Sprintf("/courses/%s/assignments/%s/submissions/%s", courseID, assignID, studentID)
	data, err := client.PUT(endpoint, form)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if jsonOutput {
		fmt.Println(string(data))
		return
	}

	var result struct {
		ID            int      `json:"id"`
		UserID        int      `json:"user_id"`
		Score         *float64 `json:"score"`
		Grade         string   `json:"grade"`
		WorkflowState string   `json:"workflow_state"`
	}
	json.Unmarshal(data, &result)

	if hasScore {
		ui.Success(fmt.Sprintf("Graded student %s", studentID))
		if result.Grade != "" {
			fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Grade:"), result.Grade)
		}
		if result.Score != nil {
			fmt.Printf("  %s  %.2f\n", ui.C(ui.Bold, "Score:"), *result.Score)
		}
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Status:"), ui.StatusColor(result.WorkflowState))
	} else {
		ui.Success(fmt.Sprintf("Comment posted to student %s's submission", studentID))
	}
	if comment != "" && hasScore {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Comment:"), comment)
	}
	fmt.Println()
}

// runAnnounce posts a course announcement.
//
// Usage:
//
//	canvas-cli announce <course_id> --title "Title" --message "Body text"
//
// Canvas models announcements as discussion topics with is_announcement=true.
func runAnnounce(args []string) {
	if len(args) < 1 {
		ui.Error("usage: canvas-cli announce <course_id> --title \"Title\" --message \"Body\"")
		os.Exit(1)
	}

	courseID := args[0]
	remaining := args[1:]

	var title, message string
	for i := 0; i < len(remaining); i++ {
		switch remaining[i] {
		case "--title":
			if i+1 < len(remaining) {
				title = remaining[i+1]
				i++
			}
		case "--message", "--body":
			if i+1 < len(remaining) {
				message = remaining[i+1]
				i++
			}
		}
	}

	if title == "" || message == "" {
		ui.Error("both --title and --message are required")
		os.Exit(1)
	}

	form := url.Values{
		"title":           {title},
		"message":         {message},
		"is_announcement": {"true"},
		"published":       {"true"},
	}

	endpoint := fmt.Sprintf("/courses/%s/discussion_topics", courseID)
	data, err := client.POST(endpoint, form)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if jsonOutput {
		fmt.Println(string(data))
		return
	}

	var result struct {
		ID       int    `json:"id"`
		Title    string `json:"title"`
		PostedAt string `json:"posted_at"`
		HTMLURL  string `json:"html_url"`
	}
	json.Unmarshal(data, &result)

	ui.Success(fmt.Sprintf("Announcement posted! (ID: %d)", result.ID))
	fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Title:"), result.Title)
	if result.PostedAt != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Posted:"), ui.FormatDate(result.PostedAt))
	}
	if result.HTMLURL != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "URL:"), result.HTMLURL)
	}
	fmt.Println()
}

// runCreateAssignment creates a new assignment in a course.
//
// Usage:
//
//	canvas-cli create-assignment <course_id> --name "Title" [opts]
//	  --points <n>            Points possible (default 100)
//	  --due <date>            Due date: "2026-09-19" or "2026-09-19T23:59:00"
//	  --description "text"    Assignment description / instructions
//	  --types "a,b"           Submission types (default online_upload,online_text_entry)
//	  --publish               Publish immediately (default: unpublished draft)
func runCreateAssignment(args []string) {
	if len(args) < 1 {
		ui.Error("usage: canvas-cli create-assignment <course_id> --name \"Title\" [--points <n>] [--due <date>] [--description \"...\"] [--publish]")
		os.Exit(1)
	}

	courseID := args[0]
	remaining := args[1:]

	var name, due, description, types string
	points := "100"
	publish := false

	for i := 0; i < len(remaining); i++ {
		switch remaining[i] {
		case "--name":
			if i+1 < len(remaining) {
				name = remaining[i+1]
				i++
			}
		case "--points":
			if i+1 < len(remaining) {
				points = remaining[i+1]
				i++
			}
		case "--due":
			if i+1 < len(remaining) {
				due = remaining[i+1]
				i++
			}
		case "--description", "--desc":
			if i+1 < len(remaining) {
				description = remaining[i+1]
				i++
			}
		case "--types":
			if i+1 < len(remaining) {
				types = remaining[i+1]
				i++
			}
		case "--publish":
			publish = true
		}
	}

	if name == "" {
		ui.Error("--name is required")
		os.Exit(1)
	}

	form := url.Values{
		"assignment[name]":            {name},
		"assignment[points_possible]": {points},
	}
	if publish {
		form.Set("assignment[published]", "true")
	} else {
		form.Set("assignment[published]", "false")
	}
	if description != "" {
		form.Set("assignment[description]", description)
	}
	if due != "" {
		form.Set("assignment[due_at]", normalizeDueDate(due))
	}

	// Submission types (Canvas expects an array).
	subTypes := []string{"online_upload", "online_text_entry"}
	if types != "" {
		subTypes = splitAndTrim(types)
	}
	for _, t := range subTypes {
		form.Add("assignment[submission_types][]", t)
	}

	endpoint := fmt.Sprintf("/courses/%s/assignments", courseID)
	data, err := client.POST(endpoint, form)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if jsonOutput {
		fmt.Println(string(data))
		return
	}

	var result struct {
		ID             int     `json:"id"`
		Name           string  `json:"name"`
		PointsPossible float64 `json:"points_possible"`
		DueAt          string  `json:"due_at"`
		Published      bool    `json:"published"`
		HTMLURL        string  `json:"html_url"`
	}
	json.Unmarshal(data, &result)

	ui.Success(fmt.Sprintf("Assignment created! (ID: %d)", result.ID))
	fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Name:"), result.Name)
	fmt.Printf("  %s  %.0f\n", ui.C(ui.Bold, "Points:"), result.PointsPossible)
	if result.DueAt != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Due:"), ui.FormatDate(result.DueAt))
	}
	state := "draft (unpublished)"
	if result.Published {
		state = "published"
	}
	fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "State:"), state)
	if result.HTMLURL != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "URL:"), result.HTMLURL)
	}
	if !result.Published {
		fmt.Printf("  %s\n", ui.C(ui.Dim, "Re-run with --publish (or edit in Canvas) to make it visible to students."))
	}
	fmt.Println()
}

// runUpload uploads a file into a course's Files area (course material).
//
// Usage:
//
//	canvas-cli upload <course_id> --file <path> [--file <path> ...]
func runUpload(args []string) {
	if len(args) < 1 {
		ui.Error("usage: canvas-cli upload <course_id> --file <path> [--file <path> ...]")
		os.Exit(1)
	}

	courseID := args[0]
	remaining := args[1:]

	var filePaths []string
	for i := 0; i < len(remaining); i++ {
		if remaining[i] == "--file" && i+1 < len(remaining) {
			filePaths = append(filePaths, remaining[i+1])
			i++
		}
	}

	if len(filePaths) == 0 {
		ui.Error("provide at least one --file <path>")
		os.Exit(1)
	}

	endpoint := fmt.Sprintf("/courses/%s/files", courseID)
	for _, path := range filePaths {
		ui.Info(fmt.Sprintf("Uploading %s ...", path))
		res, err := client.UploadFile(endpoint, path)
		if err != nil {
			ui.Error(fmt.Sprintf("uploading %s: %s", path, err.Error()))
			os.Exit(1)
		}
		ui.Success(fmt.Sprintf("Uploaded %s (file ID: %d, %d bytes)", res.DisplayName, res.ID, res.Size))
	}
	fmt.Println()
	fmt.Printf("  %s\n", ui.C(ui.Dim, "Files uploaded to the course Files area. View with: canvas-cli files "+courseID))
	fmt.Println()
}

// normalizeDueDate accepts a plain date ("2026-09-19") or a full timestamp and
// returns an ISO 8601 value Canvas accepts. A bare date is treated as 23:59
// local time on that day.
func normalizeDueDate(due string) string {
	if strings.Contains(due, "T") {
		return due
	}
	if t, err := time.ParseInLocation("2006-01-02", due, time.Local); err == nil {
		return t.Add(23*time.Hour + 59*time.Minute).Format(time.RFC3339)
	}
	return due
}

// splitAndTrim splits a comma-separated list and trims whitespace from each item.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
