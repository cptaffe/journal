package main

import (
	"fmt"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Generate hash
func hash(f *os.File) ([64]byte, error) {
	if _, err := f.Seek(0, 0); err != nil {
		return [64]byte{}, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return [64]byte{}, err
	}
	return sha3.Sum512(b), nil
}

func main() {
	home, ok := os.LookupEnv("HOME")
	if !ok {
		log.Fatalf("$HOME not set")
	}
	visual, ok := os.LookupEnv("EDITOR")
	if !ok {
		log.Fatalf("$EDITOR not set")
	}
	year, month, day := time.Now().Date()
	jf := fmt.Sprintf("%s/journal/%d.%d.%d.md", home, day, month, year)
	var f *os.File
	if _, err := os.Stat(jf); os.IsNotExist(err) {
		f, err = os.Create(jf)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(f, "%s\n\n", time.Now().Format(time.UnixDate))
	} else if err == nil {
		f, err = os.OpenFile(jf, os.O_RDWR, 0644)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}
	ohash, err := hash(f)
	if err != nil {
		log.Fatal(err)
	}
	// Open editor
	cmd := exec.Command(visual, f.Name())
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	// Generate post-edit hash
	nhash, err := hash(f)
	// If hashes are different, ask for commit message
	if ohash != nhash {
		// Add file to git
		cmd := exec.Command("git", "-C", filepath.Dir(f.Name()), "add", f.Name())
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		// Read commit message from user
		fmt.Printf("Enter a summary and then press ^D\n\n")
		// Commit file to git
		cmd = exec.Command("git", "-C", filepath.Dir(f.Name()), "commit", "--file", "-")
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}
}
