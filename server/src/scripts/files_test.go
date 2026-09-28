package scripts

import "testing"

func TestPrepareFileInputsDefaultMain(t *testing.T) {
	files, err := PrepareFileInputs("javascript", "console.log(1)", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Name != "main.js" || files[0].Kind != FileKindMain {
		t.Fatalf("unexpected files: %+v", files)
	}
	if files[0].Content != "console.log(1)" {
		t.Fatalf("content = %q", files[0].Content)
	}
}

func TestPrepareFileInputsRejectsBadName(t *testing.T) {
	_, err := PrepareFileInputs("javascript", "", []FileInput{
		{Name: "main.js", Kind: FileKindMain},
		{Name: "../x.js", Kind: FileKindLocal},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPrepareFileInputsRequiresMain(t *testing.T) {
	_, err := PrepareFileInputs("python", "", []FileInput{
		{Name: "helper.py", Kind: FileKindLocal},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMainFileName(t *testing.T) {
	name, err := MainFileName("python")
	if err != nil || name != "main.py" {
		t.Fatalf("got %q %v", name, err)
	}
}
