package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// compressAndEncrypt compresses and encrypts files from a directory
func compressAndEncrypt(inputDir, outputFile, passFile, saltFile string) error {
	// Command for compressing
	tarCmd := exec.Command("tar", "-cJ", inputDir)

	// Command for encrypting
	opensslCmd := exec.Command("openssl", "enc", "-aes-256-cbc", "-salt", "-pbkdf2", "-out", outputFile+".enc", "-pass", "file:"+passFile, "-pass", "file:"+saltFile)

	// Create a pipe to connect the tar command's output to the openssl command's input
	pipe, err := tarCmd.StdoutPipe()
	if err != nil {
		return err
	}
	opensslCmd.Stdin = pipe

	// Capture the tar command's error output
	tarError := &bytes.Buffer{}
	tarCmd.Stderr = tarError

	// Start the openssl command
	if err := opensslCmd.Start(); err != nil {
		return err
	}

	// Start the tar command
	if err := tarCmd.Start(); err != nil {
		return err
	}

	// Wait for the tar command to finish
	if err := tarCmd.Wait(); err != nil {
		return errors.New(err.Error() + ": " + tarError.String())
	}

	// Wait for the openssl command to finish
	if err := opensslCmd.Wait(); err != nil {
		return err
	}

	return nil
}

// decryptAndDecompress decrypts and decompresses a file into a specific directory
func decryptAndDecompress(inputFile, outputDir, passFile, saltFile string) error {
	// Command for decrypting
	opensslCmd := exec.Command("openssl", "enc", "-d", "-aes-256-cbc", "-pbkdf2", "-in", inputFile, "-pass", "file:"+passFile, "-pass", "file:"+saltFile)

	// Command for decompressing
	tarCmd := exec.Command("tar", "-xJf", "-", "-C", outputDir)

	// Create a pipe to connect the openssl command's output to the tar command's input
	pipe, err := opensslCmd.StdoutPipe()
	if err != nil {
		return err
	}
	tarCmd.Stdin = pipe

	// Capture the openssl command's error output
	opensslError := &bytes.Buffer{}
	opensslCmd.Stderr = opensslError

	// Start the openssl command
	if err := opensslCmd.Start(); err != nil {
		return err
	}

	// Run the tar command
	if err := tarCmd.Run(); err != nil {
		return errors.New(err.Error() + ": " + opensslError.String())
	}

	// Wait for the openssl command to finish
	if err := opensslCmd.Wait(); err != nil {
		return err
	}

	return nil
}

// sendToBackblaze sends a file to Backblaze using rclone
func sendToBackblaze(sourceFile, remoteDestination string) error {
	cmd := exec.Command("rclone", "copy", sourceFile, remoteDestination)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [option] [arguments]")
		fmt.Println("Available options:")
		fmt.Println("  compress_encrypt [inputDir] [outputFile] [passFile] [saltFile]")
		fmt.Println("  decrypt_decompress [inputFile] [outputDir] [passFile] [saltFile]")
		fmt.Println("  send_to_backblaze [sourceFile] [remoteDestination]")
		return
	}

	switch os.Args[1] {
	case "compress_encrypt":
		if len(os.Args) != 6 {
			fmt.Println("Incorrect number of arguments. Usage: go run main.go compress_encrypt [inputDir] [outputFile] [passFile] [saltFile]")
			return
		}
		err := compressAndEncrypt(os.Args[2], os.Args[3], os.Args[4], os.Args[5])
		if err != nil {
			fmt.Println("Error compressing and encrypting:", err)
			return
		}
	case "decrypt_decompress":
		if len(os.Args) != 6 {
			fmt.Println("Incorrect number of arguments. Usage: go run main.go decrypt_decompress [inputFile] [outputDir] [passFile] [saltFile]")
			return
		}
		err := decryptAndDecompress(os.Args[2], os.Args[3], os.Args[4], os.Args[5])
		if err != nil {
			fmt.Println("Error decrypting and decompressing:", err)
			return
		}
	case "send_to_backblaze":
		if len(os.Args) != 4 {
			fmt.Println("Incorrect number of arguments. Usage: go run main.go send_to_backblaze [sourceFile] [remoteDestination]")
			return
		}
		err := sendToBackblaze(os.Args[2], os.Args[3])
		if err != nil {
			fmt.Println("Error sending to Backblaze:", err)
			return
		}
	default:
		fmt.Println("Invalid option:", os.Args[1])
		return
	}

	fmt.Println("Operation completed successfully.")
}
