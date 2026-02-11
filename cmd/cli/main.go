package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/1it/go-submission-service/internal/config"
	"github.com/1it/go-submission-service/internal/database"
	"github.com/1it/go-submission-service/internal/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// Logging / usage helpers
// ─────────────────────────────────────────────────────────────────────────────
func setupLogging() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}

func printUsage() {
	fmt.Println("Form Submissions Management CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  list-subscribers [--status=<status>]               List subscribers (legacy table)")
	fmt.Println("  list-submissions [--status=<status>] [--form-type=<type>]  List form submissions")
	fmt.Println("  stats-subscribers                                   Show subscriber statistics (legacy)")
	fmt.Println("  stats-submissions                                   Show submission statistics")
	fmt.Println("  remove-subscriber <email>                           Remove a subscriber by email (legacy)")
	fmt.Println("  remove-submission <email>                           Remove a submission by email")
	fmt.Println("  config-docs                                         Print configuration documentation")
	fmt.Println("  migrate | migrate-status | migrate-to <v>           DB migration helpers")

	fmt.Println("\nFlags:")
	fmt.Println("  --database-path  Path to SQLite DB (defaults to config value)")
	fmt.Println("  --status         Filter by status (processed, pending, …)")
	fmt.Println("  --form-type      Filter submissions by form_type (business_contact, …)")
}

// ─────────────────────────────────────────────────────────────────────────────
// main
// ─────────────────────────────────────────────────────────────────────────────
func main() {
	setupLogging()

	// CLI flags
	dbPathFlag := flag.String("database-path", "", "Path to the SQLite database file")
	statusFlag := flag.String("status", "", "Filter by status")
	formTypeFlag := flag.String("form-type", "", "Filter submissions by form type")
	flag.Parse()

	// Command (first arg after flags)
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}
	cmd := args[0]

	// Load config to obtain default DB path
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}
	dbPath := *dbPathFlag
	if dbPath == "" {
		dbPath = cfg.Database.Path
	}

	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	repo := db.GetRepository()

	// Dispatch
	switch cmd {
	// ───────── LIST ────────────────────────────────────────────────
	case "list", "list-subscribers":
		listSubscribers(repo, *statusFlag)
	case "list-submissions":
		listSubmissions(repo, *statusFlag, *formTypeFlag)

	// ───────── STATS ───────────────────────────────────────────────
	case "stats", "stats-subscribers":
		showSubscriberStats(repo)
	case "stats-submissions":
		showSubmissionStats(repo)

	// ───────── REMOVE ──────────────────────────────────────────────
	case "remove", "remove-subscriber":
		if len(args) < 2 {
			log.Fatal("Email required")
		}
		removeSubscriber(repo, args[1])
	case "remove-submission":
		if len(args) < 2 {
			log.Fatal("Email required")
		}
		removeSubmission(repo, args[1])

	// ───────── MISC ────────────────────────────────────────────────
	case "config-docs":
		fmt.Println(config.GenerateConfigDocs())
	case "migrate":
		runMigrations(db)
	case "migrate-status":
		showMigrationStatus(db)
	case "migrate-to":
		if len(args) < 2 {
			log.Fatal("Version required")
		}
		migrateToVersion(db, args[1])

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Subscriber helpers (legacy table)
// ─────────────────────────────────────────────────────────────────────────────
func listSubscribers(repo *database.Repository, status string) {
	subs, err := repo.Subscribers.List(status)
	if err != nil {
		log.Fatalf("list subscribers: %v", err)
	}
	if len(subs) == 0 {
		log.Println("No subscribers found")
		return
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tEmail\tStatus\tCreated\tUpdated")
	fmt.Fprintln(tw, "--\t-----\t------\t-------\t-------")
	for _, s := range subs {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n",
			s.ID, s.Email, s.Status,
			s.CreatedAt.Format(time.RFC3339),
			s.UpdatedAt.Format(time.RFC3339))
	}
	tw.Flush()
	fmt.Printf("\nTotal: %d subscribers\n", len(subs))
}

func showSubscriberStats(repo *database.Repository) {
	stats, err := repo.Subscribers.GetStats()
	if err != nil {
		log.Fatalf("stats subscribers: %v", err)
	}
	printStats(stats)
}

func removeSubscriber(repo *database.Repository, email string) {
	sub, err := repo.Subscribers.GetByEmail(email)
	if err != nil {
		log.Fatalf("subscriber not found: %s", email)
	}
	fmt.Printf("Remove subscriber %s (status %s)? [y/N]: ", sub.Email, sub.Status)
	var c string
	if _, err := fmt.Scanln(&c); err != nil {
		log.Printf("Error reading input: %v", err)
		return
	}
	if c != "y" && c != "Y" {
		return
	}
	if err := repo.Subscribers.Remove(email); err != nil {
		log.Fatalf("remove subscriber: %v", err)
	}
	log.Printf("Removed subscriber %s", email)
}

// ─────────────────────────────────────────────────────────────────────────────
// Submission helpers (main table)
// ─────────────────────────────────────────────────────────────────────────────
func listSubmissions(repo *database.Repository, status, formType string) {
	filters := database.SubmissionFilters{Status: status, FormType: formType}
	subs, err := repo.Submissions.List(filters)
	if err != nil {
		log.Fatalf("list submissions: %v", err)
	}
	if len(subs) == 0 {
		log.Println("No submissions found")
		return
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tEmail\tStatus\tFormType\tCreated\tUpdated")
	fmt.Fprintln(tw, "--\t-----\t------\t--------\t-------\t-------")
	for _, s := range subs {
		email := ""
		if s.FormData != nil {
			if v, ok := s.FormData["work_email"].(string); ok {
				email = v
			} else if v, ok := s.FormData["email"].(string); ok {
				email = v
			}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			s.ID, email, s.Status, s.FormType,
			s.CreatedAt.Format(time.RFC3339),
			s.UpdatedAt.Format(time.RFC3339))
	}
	tw.Flush()
	fmt.Printf("\nTotal: %d submissions\n", len(subs))
}

func showSubmissionStats(repo *database.Repository) {
	stats, err := repo.Submissions.GetStats()
	if err != nil {
		log.Fatalf("stats submissions: %v", err)
	}
	printStats(stats)
}

func removeSubmission(repo *database.Repository, email string) {
	sub, err := findSubmissionByEmail(repo, email)
	if err != nil {
		log.Fatalf("submission not found: %s", email)
	}
	dispEmail := email
	if sub.FormData != nil {
		if v, ok := sub.FormData["work_email"].(string); ok {
			dispEmail = v
		} else if v, ok := sub.FormData["email"].(string); ok {
			dispEmail = v
		}
	}
	fmt.Printf("Remove submission %s (status %s, form_type %s)? [y/N]: ", dispEmail, sub.Status, sub.FormType)
	var c string
	if _, err := fmt.Scanln(&c); err != nil {
		log.Printf("Error reading input: %v", err)
		return
	}
	if c != "y" && c != "Y" {
		return
	}
	if err := repo.Submissions.Delete(sub.ID); err != nil {
		log.Fatalf("delete submission: %v", err)
	}
	log.Printf("Removed submission %s", dispEmail)
}

// helper to search by email across possible fields
func findSubmissionByEmail(repo *database.Repository, email string) (*models.Submission, error) {
	subs, err := repo.Submissions.List(database.SubmissionFilters{})
	if err != nil {
		return nil, err
	}
	for _, s := range subs {
		if s.FormData == nil {
			continue
		}
		for _, field := range []string{"work_email", "email", "user_email", "contact_email"} {
			if v, ok := s.FormData[field].(string); ok && v == email {
				return s, nil
			}
		}
	}
	return nil, fmt.Errorf("not found")
}

// ─────────────────────────────────────────────────────────────────────────────
// Generic helpers
// ─────────────────────────────────────────────────────────────────────────────
func printStats(stats map[string]int) {
	if len(stats) == 0 {
		log.Println("No statistics available")
		return
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Status\tCount")
	fmt.Fprintln(tw, "------\t-----")
	total := stats["total"]
	delete(stats, "total")
	for k, v := range stats {
		fmt.Fprintf(tw, "%s\t%d\n", k, v)
	}
	fmt.Fprintln(tw, "------\t-----")
	fmt.Fprintf(tw, "Total\t%d\n", total)
	tw.Flush()
}

// ─────────────────────────────────────────────────────────────────────────────
// Migration helpers (kept from original file)
// ─────────────────────────────────────────────────────────────────────────────
func runMigrations(db *database.SQLiteDB) {
	log.Println("Running database migrations…")
	if err := db.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("Migrations completed successfully")
}

func showMigrationStatus(db *database.SQLiteDB) {
	status, err := db.GetMigrationStatus()
	if err != nil {
		log.Fatalf("migration status: %v", err)
	}
	fmt.Println("Migration Status:")
	fmt.Printf("Current Version: %d\n", status["current_version"])
	fmt.Printf("Latest  Version: %d\n", status["latest_version"])
	fmt.Printf("Applied Migrations: %d\n", status["applied_count"])
	fmt.Printf("Pending Migrations: %d\n", status["pending_count"])
	if status["pending_count"].(int) > 0 {
		fmt.Println("Run './cli migrate' to apply pending migrations")
	}
}

func migrateToVersion(db *database.SQLiteDB, versionStr string) {
	var ver int
	if _, err := fmt.Sscanf(versionStr, "%d", &ver); err != nil {
		log.Fatalf("invalid version: %s", versionStr)
	}
	log.Printf("Migrating to version %d…", ver)
	if err := db.MigrateToVersion(ver); err != nil {
		log.Fatalf("migrate to version: %v", err)
	}
	log.Printf("Successfully migrated to version %d", ver)
}
