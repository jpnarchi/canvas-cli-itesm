package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"canvas-cli/internal/ui"
)

func runSubmissions(args []string) {
	if len(args) < 2 {
		ui.Error("usage: canvas-cli submissions <course_id> <assignment_id> [--all | --student <id>]")
		os.Exit(1)
	}

	courseID := args[0]
	assignID := args[1]
	remaining := args[2:]

	// Teacher views: --all lists every student's submission for the assignment,
	// --student <id> shows one specific student's submission.
	for i := 0; i < len(remaining); i++ {
		switch remaining[i] {
		case "--all":
			listAllSubmissions(courseID, assignID)
			return
		case "--student":
			if i+1 < len(remaining) {
				showStudentSubmission(courseID, assignID, remaining[i+1])
				return
			}
			ui.Error("--student requires a student ID")
			os.Exit(1)
		}
	}

	data, err := client.GET(fmt.Sprintf("/courses/%s/assignments/%s/submissions/self?include[]=submission_comments&include[]=rubric_assessment", courseID, assignID))
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if jsonOutput {
		fmt.Println(string(data))
		return
	}

	var sub struct {
		ID            int      `json:"id"`
		Score         *float64 `json:"score"`
		Grade         string   `json:"grade"`
		WorkflowState string   `json:"workflow_state"`
		SubmittedAt   string   `json:"submitted_at"`
		Late          bool     `json:"late"`
		Missing       bool     `json:"missing"`
		Attempt       int      `json:"attempt"`
		Body          string   `json:"body"`
		URL           string   `json:"url"`
		PreviewURL    string   `json:"preview_url"`
		Attachments   []struct {
			ID          int    `json:"id"`
			DisplayName string `json:"display_name"`
			URL         string `json:"url"`
			Size        int    `json:"size"`
		} `json:"attachments"`
		SubmissionComments []struct {
			ID         int    `json:"id"`
			AuthorName string `json:"author_name"`
			Comment    string `json:"comment"`
			CreatedAt  string `json:"created_at"`
		} `json:"submission_comments"`
	}
	if err := json.Unmarshal(data, &sub); err != nil {
		ui.Error("parsing submission: " + err.Error())
		os.Exit(1)
	}

	ui.Header("Your Submission")
	fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Status:"), ui.StatusColor(sub.WorkflowState))
	if sub.SubmittedAt != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Submitted:"), ui.FormatDate(sub.SubmittedAt))
	}
	if sub.Score != nil {
		fmt.Printf("  %s  %.1f\n", ui.C(ui.Bold, "Score:"), *sub.Score)
	}
	if sub.Grade != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Grade:"), sub.Grade)
	}
	if sub.Attempt > 0 {
		fmt.Printf("  %s  %d\n", ui.C(ui.Bold, "Attempt:"), sub.Attempt)
	}
	if sub.Late {
		fmt.Printf("  %s\n", ui.C(ui.Red, "  Late submission"))
	}
	if sub.Missing {
		fmt.Printf("  %s\n", ui.C(ui.Red, "  Missing"))
	}
	if sub.Body != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Body:"), ui.Truncate(sub.Body, 200))
	}
	if sub.URL != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "URL:"), sub.URL)
	}

	if len(sub.Attachments) > 0 {
		fmt.Println()
		fmt.Printf("  %s\n", ui.C(ui.Bold+ui.Cyan, "Attachments"))
		for _, att := range sub.Attachments {
			fmt.Printf("    • %s (ID: %d, %d bytes)\n", att.DisplayName, att.ID, att.Size)
		}
	}

	if len(sub.SubmissionComments) > 0 {
		fmt.Println()
		fmt.Printf("  %s\n", ui.C(ui.Bold+ui.Cyan, "Comments"))
		for _, c := range sub.SubmissionComments {
			fmt.Printf("    %s %s\n", ui.C(ui.Bold, c.AuthorName), ui.C(ui.Dim, ui.FormatDate(c.CreatedAt)))
			fmt.Printf("    %s\n\n", c.Comment)
		}
	}
	fmt.Println()
}

// listAllSubmissions shows every student's submission for an assignment
// (teacher view). It surfaces who has submitted, who is missing, and current
// scores so a grader knows what still needs attention.
func listAllSubmissions(courseID, assignID string) {
	items, err := client.GetPaginated(fmt.Sprintf("/courses/%s/assignments/%s/submissions?include[]=user", courseID, assignID))
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if jsonOutput {
		out, _ := json.Marshal(items)
		fmt.Println(string(out))
		return
	}

	type submission struct {
		UserID        int      `json:"user_id"`
		Score         *float64 `json:"score"`
		Grade         string   `json:"grade"`
		WorkflowState string   `json:"workflow_state"`
		SubmittedAt   string   `json:"submitted_at"`
		Late          bool     `json:"late"`
		Missing       bool     `json:"missing"`
		User          struct {
			Name string `json:"name"`
		} `json:"user"`
	}

	ui.Header(fmt.Sprintf("Submissions — Course %s / Assignment %s", courseID, assignID))

	rows := make([][]string, 0, len(items))
	for _, raw := range items {
		var s submission
		if err := json.Unmarshal(raw, &s); err != nil {
			continue
		}
		status := s.WorkflowState
		switch {
		case s.Missing:
			status = ui.StatusColor("missing")
		case s.WorkflowState == "graded":
			status = ui.StatusColor("graded")
		case s.WorkflowState == "submitted":
			status = ui.StatusColor("submitted")
		case s.WorkflowState == "unsubmitted":
			status = ui.StatusColor("unsubmitted")
		}
		if s.Late {
			status += ui.C(ui.Red, " (late)")
		}
		score := ""
		if s.Score != nil {
			score = fmt.Sprintf("%.1f", *s.Score)
		}
		if s.Grade != "" {
			score = s.Grade
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", s.UserID),
			ui.Truncate(s.User.Name, 30),
			status,
			score,
			ui.FormatDate(s.SubmittedAt),
		})
	}

	ui.Table([]string{"STUDENT_ID", "NAME", "STATUS", "SCORE", "SUBMITTED"}, rows)
	fmt.Println()
	fmt.Printf("  %s\n", ui.C(ui.Dim, "Grade with: canvas-cli grade "+courseID+" "+assignID+" <student_id> <score> [--comment \"...\"]"))
	fmt.Println()
}

// showStudentSubmission shows one student's submission in detail (teacher view),
// including any file attachments and comments so the grader can review the work.
func showStudentSubmission(courseID, assignID, studentID string) {
	data, err := client.GET(fmt.Sprintf("/courses/%s/assignments/%s/submissions/%s?include[]=submission_comments&include[]=user", courseID, assignID, studentID))
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if jsonOutput {
		fmt.Println(string(data))
		return
	}

	var sub struct {
		UserID        int      `json:"user_id"`
		Score         *float64 `json:"score"`
		Grade         string   `json:"grade"`
		WorkflowState string   `json:"workflow_state"`
		SubmittedAt   string   `json:"submitted_at"`
		Late          bool     `json:"late"`
		Missing       bool     `json:"missing"`
		Attempt       int      `json:"attempt"`
		Body          string   `json:"body"`
		URL           string   `json:"url"`
		User          struct {
			Name string `json:"name"`
		} `json:"user"`
		Attachments []struct {
			ID          int    `json:"id"`
			DisplayName string `json:"display_name"`
			URL         string `json:"url"`
			Size        int    `json:"size"`
		} `json:"attachments"`
		SubmissionComments []struct {
			AuthorName string `json:"author_name"`
			Comment    string `json:"comment"`
			CreatedAt  string `json:"created_at"`
		} `json:"submission_comments"`
	}
	if err := json.Unmarshal(data, &sub); err != nil {
		ui.Error("parsing submission: " + err.Error())
		os.Exit(1)
	}

	title := sub.User.Name
	if title == "" {
		title = "Student " + studentID
	}
	ui.Header(title + "'s Submission")
	fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Status:"), ui.StatusColor(sub.WorkflowState))
	if sub.SubmittedAt != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Submitted:"), ui.FormatDate(sub.SubmittedAt))
	}
	if sub.Score != nil {
		fmt.Printf("  %s  %.2f\n", ui.C(ui.Bold, "Score:"), *sub.Score)
	}
	if sub.Grade != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Grade:"), sub.Grade)
	}
	if sub.Late {
		fmt.Printf("  %s\n", ui.C(ui.Red, "  Late submission"))
	}
	if sub.Missing {
		fmt.Printf("  %s\n", ui.C(ui.Red, "  Missing"))
	}
	if sub.Body != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Body:"), ui.Truncate(sub.Body, 300))
	}
	if sub.URL != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "URL:"), sub.URL)
	}

	if len(sub.Attachments) > 0 {
		fmt.Println()
		fmt.Printf("  %s\n", ui.C(ui.Bold+ui.Cyan, "Attachments"))
		for _, att := range sub.Attachments {
			fmt.Printf("    • %s (ID: %d, %d bytes)\n", att.DisplayName, att.ID, att.Size)
		}
		fmt.Printf("  %s\n", ui.C(ui.Dim, "Download with: canvas-cli download <file_id>"))
	}

	if len(sub.SubmissionComments) > 0 {
		fmt.Println()
		fmt.Printf("  %s\n", ui.C(ui.Bold+ui.Cyan, "Comments"))
		for _, c := range sub.SubmissionComments {
			fmt.Printf("    %s %s\n", ui.C(ui.Bold, c.AuthorName), ui.C(ui.Dim, ui.FormatDate(c.CreatedAt)))
			fmt.Printf("    %s\n\n", c.Comment)
		}
	}
	fmt.Println()
	fmt.Printf("  %s\n", ui.C(ui.Dim, "Grade with: canvas-cli grade "+courseID+" "+assignID+" "+studentID+" <score> [--comment \"...\"]"))
	fmt.Println()
}

func runSubmit(args []string) {
	if len(args) < 2 {
		ui.Error("usage: canvas-cli submit <course_id> <assignment_id> --text \"content\" | --url <url> | --file <path> [--dry-run]")
		os.Exit(1)
	}

	courseID := args[0]
	assignID := args[1]
	remaining := args[2:]

	var submissionBody, submissionURL string
	var filePaths []string
	var hasText, hasURL, dryRun bool

	for i := 0; i < len(remaining); i++ {
		switch remaining[i] {
		case "--text":
			if i+1 < len(remaining) {
				hasText = true
				submissionBody = remaining[i+1]
				i++
			}
		case "--url":
			if i+1 < len(remaining) {
				hasURL = true
				submissionURL = remaining[i+1]
				i++
			}
		case "--file":
			if i+1 < len(remaining) {
				filePaths = append(filePaths, remaining[i+1])
				i++
			} else {
				ui.Error("--file requires a path")
				os.Exit(1)
			}
		case "--dry-run":
			dryRun = true
		}
	}

	// Determine submission type. File uploads take precedence and can be
	// validated without actually submitting via --dry-run.
	if len(filePaths) > 0 {
		if hasText || hasURL {
			ui.Error("--file cannot be combined with --text or --url")
			os.Exit(1)
		}
		runFileSubmit(courseID, assignID, filePaths, dryRun)
		return
	}

	var submissionType string
	switch {
	case hasText:
		submissionType = "online_text_entry"
	case hasURL:
		submissionType = "online_url"
	default:
		ui.Error("specify submission type: --text \"content\", --url <url>, or --file <path>")
		os.Exit(1)
	}

	if dryRun {
		ui.Warning("--dry-run only applies to --file uploads; nothing was submitted.")
		return
	}

	form := url.Values{
		"submission[submission_type]": {submissionType},
	}
	if submissionBody != "" {
		form.Set("submission[body]", submissionBody)
	}
	if submissionURL != "" {
		form.Set("submission[url]", submissionURL)
	}

	endpoint := fmt.Sprintf("/courses/%s/assignments/%s/submissions", courseID, assignID)
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
		ID            int    `json:"id"`
		WorkflowState string `json:"workflow_state"`
		SubmittedAt   string `json:"submitted_at"`
	}
	json.Unmarshal(data, &result)

	ui.Success(fmt.Sprintf("Submitted! (ID: %d)", result.ID))
	fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Status:"), ui.StatusColor(result.WorkflowState))
	if result.SubmittedAt != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Time:"), ui.FormatDate(result.SubmittedAt))
	}

	fmt.Println()
}

// runFileSubmit uploads one or more attachments to a submission. Each file is
// pushed into the user's Canvas storage first (via the assignment's file-upload
// endpoint). When dryRun is true the flow stops after the uploads succeed and
// the assignment is NOT submitted — this validates that attachments upload
// correctly without turning anything in.
func runFileSubmit(courseID, assignID string, filePaths []string, dryRun bool) {
	uploadEndpoint := fmt.Sprintf("/courses/%s/assignments/%s/submissions/self/files", courseID, assignID)

	fileIDs := make([]int, 0, len(filePaths))
	for _, path := range filePaths {
		ui.Info(fmt.Sprintf("Uploading %s ...", path))
		res, err := client.UploadFile(uploadEndpoint, path)
		if err != nil {
			ui.Error(fmt.Sprintf("uploading %s: %s", path, err.Error()))
			os.Exit(1)
		}
		ui.Success(fmt.Sprintf("Uploaded %s (file ID: %d)", res.DisplayName, res.ID))
		fileIDs = append(fileIDs, res.ID)
	}

	if dryRun {
		fmt.Println()
		ui.Warning("Dry run — files uploaded to Canvas but the assignment was NOT submitted.")
		fmt.Printf("  %s  %v\n", ui.C(ui.Bold, "Uploaded file IDs:"), fileIDs)
		fmt.Printf("  %s\n", ui.C(ui.Dim, "Re-run without --dry-run to actually submit these files."))
		fmt.Println()
		return
	}

	form := url.Values{
		"submission[submission_type]": {"online_upload"},
	}
	for _, id := range fileIDs {
		form.Add("submission[file_ids][]", fmt.Sprintf("%d", id))
	}

	endpoint := fmt.Sprintf("/courses/%s/assignments/%s/submissions", courseID, assignID)
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
		ID            int    `json:"id"`
		WorkflowState string `json:"workflow_state"`
		SubmittedAt   string `json:"submitted_at"`
	}
	json.Unmarshal(data, &result)

	ui.Success(fmt.Sprintf("Submitted %d file(s)! (ID: %d)", len(fileIDs), result.ID))
	fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Status:"), ui.StatusColor(result.WorkflowState))
	if result.SubmittedAt != "" {
		fmt.Printf("  %s  %s\n", ui.C(ui.Bold, "Time:"), ui.FormatDate(result.SubmittedAt))
	}
	fmt.Println()
}
