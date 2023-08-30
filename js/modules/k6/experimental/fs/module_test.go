package fs

import (
	"fmt"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.k6.io/k6/js/compiler"
	"go.k6.io/k6/js/modulestest"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/lib/fsext"
	"go.k6.io/k6/metrics"
)

func TestOpen(t *testing.T) {
	t.Parallel()

	t.Run("opening existing file should succeed", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			openPath string
			wantPath string
		}{
			{
				name:     "open absolute path",
				openPath: fsext.FilePathSeparator + "bonjour.txt",
				wantPath: fsext.FilePathSeparator + "bonjour.txt",
			},
			{
				name:     "open relative path",
				openPath: filepath.Join(".", fsext.FilePathSeparator, "bonjour.txt"),
				wantPath: fsext.FilePathSeparator + "bonjour.txt",
			},
			{
				name:     "open path with ..",
				openPath: fsext.FilePathSeparator + "dir" + fsext.FilePathSeparator + ".." + fsext.FilePathSeparator + "bonjour.txt",
				wantPath: fsext.FilePathSeparator + "bonjour.txt",
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				runtime, err := newConfiguredRuntime(t)
				require.NoError(t, err)

				fs := newTestFs(t, func(fs afero.Fs) error {
					fileErr := afero.WriteFile(fs, tt.wantPath, []byte("Bonjour, le monde"), 0o644)
					if fileErr != nil {
						return fileErr
					}

					return fs.Mkdir(fsext.FilePathSeparator+"dir", 0o644)
				})
				runtime.VU.InitEnvField.FileSystems["file"] = fs
				runtime.VU.InitEnvField.CWD = &url.URL{Scheme: "file", Path: fsext.FilePathSeparator}

				_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(fmt.Sprintf(`
					let file;
					try {
						file = await fs.open(%q)
					} catch (err) {
						throw "unexpected error: " + err
					}

					if (file.path !== %q) {
						throw 'unexpected file path ' + file.path + '; expected %q';
					}
				`, tt.openPath, tt.wantPath, tt.wantPath)))

				assert.NoError(t, err)
			})
		}
	})

	t.Run("opening file in VU context should fail", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		runtime.MoveToVUContext(&lib.State{
			Tags: lib.NewVUStateTags(metrics.NewRegistry().RootTagSet().With("tag-vu", "mytag")),
		})

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(`
			try {
				const file = await fs.open('bonjour.txt')	
				throw 'unexpected promise resolution with result: ' + file;
			} catch (err) {
				if (err.name !== 'ForbiddenError') {
					throw 'unexpected error: ' + err
				}
			}

		`))

		assert.NoError(t, err)
	})

	t.Run("calling open without providing a path should fail", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(`
		let file;

		try {
			file = await fs.open()
			throw 'unexpected promise resolution with result: ' + file;
		} catch (err) {
			if (err.name !== 'TypeError') {
				throw 'unexpected error: ' + err
			}
		}

		try {
			file = await fs.open(null)
			throw 'unexpected promise resolution with result: ' + file;
		} catch (err) {
			if (err.name !== 'TypeError') {
				throw 'unexpected error: ' + err
			}
		}

		try {
			file = await fs.open(undefined)
			throw 'unexpected promise resolution with result: ' + file;
		} catch (err) {
			if (err.name !== 'TypeError') {
				throw 'unexpected error: ' + err
			}
		}
		`))

		assert.NoError(t, err)
	})

	t.Run("opening directory should fail", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		testDirPath := fsext.FilePathSeparator + "dir"
		fs := newTestFs(t, func(fs afero.Fs) error {
			return fs.Mkdir(testDirPath, 0o644)
		})

		runtime.VU.InitEnvField.FileSystems["file"] = fs

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(fmt.Sprintf(`
			try {
				const file = await fs.open(%q)
				throw 'unexpected promise resolution with result: ' + res
			} catch (err) {
				if (err.name !== 'InvalidResourceError') {
					throw 'unexpected error: ' + err
				}
			}
		`, testDirPath)))

		assert.NoError(t, err)
	})

	t.Run("opening non existing file should fail", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(`
			try {
				const file = await fs.open('doesnotexist.txt')
				throw 'unexpected promise resolution with result: ' + res
			} catch (err) {
				if (err.name !== 'NotFoundError') {
					throw 'unexpected error: ' + err
				}
			}
		`))

		assert.NoError(t, err)
	})
}

func TestFile(t *testing.T) {
	t.Parallel()

	t.Run("stat method should succeed", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		testFilePath := fsext.FilePathSeparator + "bonjour.txt"
		fs := newTestFs(t, func(fs afero.Fs) error {
			return afero.WriteFile(fs, testFilePath, []byte("Bonjour, le monde"), 0o644)
		})
		runtime.VU.InitEnvField.FileSystems["file"] = fs

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(fmt.Sprintf(`
			const file = await fs.open(%q)
			const info = await file.stat()

			if (info.name !== 'bonjour.txt') {
				throw 'unexpected file name ' + info.name + '; expected \'bonjour.txt\'';
			}

			if (info.size !== 17) {
				throw 'unexpected file size ' + info.size + '; expected 17';
			}
		`, testFilePath)))

		assert.NoError(t, err)
	})

	t.Run("stat on closed file should fail", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		testFilePath := fsext.FilePathSeparator + "bonjour.txt"
		fs := newTestFs(t, func(fs afero.Fs) error {
			return afero.WriteFile(fs, testFilePath, []byte("Bonjour, le monde"), 0o644)
		})
		runtime.VU.InitEnvField.FileSystems["file"] = fs

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(fmt.Sprintf(`
			const file = await fs.open(%q)
			await file.close()

			try {
				const info = await file.stat()
				throw 'unexpected promise resolution with result: ' + info;
			} catch (err) {
				if (err.name !== 'BadResourceError') {
					throw 'unexpected error: ' + err
				}
			}
		`, testFilePath)))

		assert.NoError(t, err)
	})

	t.Run("read in multiple iterations", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		testFilePath := fsext.FilePathSeparator + "bonjour.txt"
		fs := newTestFs(t, func(fs afero.Fs) error {
			return afero.WriteFile(fs, testFilePath, []byte("hello"), 0o644)
		})
		runtime.VU.InitEnvField.FileSystems["file"] = fs

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(fmt.Sprintf(`
			const file = await fs.open(%q);

			let fileContent = new Uint8Array(5);

			let bytesRead;
			let buffer = new Uint8Array(3);

			bytesRead = await file.read(buffer)
			if (bytesRead !== 3) {
				throw 'expected read to return 3, got ' + bytesRead + ' instead';
			}

			fileContent.set(buffer, 0);

			bytesRead = await file.read(buffer)
			if (bytesRead !== 2) {
				throw 'expected read to return 2, got ' + bytesRead + ' instead';
			}

			fileContent.set(buffer.subarray(0, bytesRead), 3);

			bytesRead = await file.read(buffer)
			if (bytesRead !== null) {
				throw 'expected read to return null, got ' + bytesRead + ' instead';
			}
		`, testFilePath)))

		assert.NoError(t, err)
	})

	t.Run("read called when end of file reached should return null and succeed", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		testFilePath := fsext.FilePathSeparator + "bonjour.txt"
		fs := newTestFs(t, func(fs afero.Fs) error {
			return afero.WriteFile(fs, testFilePath, []byte("hello"), 0o644)
		})
		runtime.VU.InitEnvField.FileSystems["file"] = fs

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(fmt.Sprintf(`
			const file = await fs.open(%q);
			let buffer = new Uint8Array(5);

			// Reading the whole file should return 5.
			let bytesRead = await file.read(buffer);
			if (bytesRead !== 5) {
				throw 'expected read to return 5, got ' + bytesRead + ' instead';
			}

			// Reading from the end of the file should return null.
			bytesRead = await file.read(buffer);
			if (bytesRead !== null) {
				throw 'expected read to return null got ' + bytesRead + ' instead';
			}
		`, testFilePath)))

		assert.NoError(t, err)
	})

	t.Run("read called with invalid argument should fail", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		testFilePath := fsext.FilePathSeparator + "bonjour.txt"
		fs := newTestFs(t, func(fs afero.Fs) error {
			return afero.WriteFile(fs, testFilePath, []byte("Bonjour, le monde"), 0o644)
		})
		runtime.VU.InitEnvField.FileSystems["file"] = fs

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(fmt.Sprintf(`
			const file = await fs.open(%q);	
			let bytesRead;

			// No argument should fail with TypeError.
			try {
				bytesRead = await file.read()
			} catch(err) {
				if (err.name !== 'TypeError') {
					throw 'unexpected error: ' + err;
				}
			}

			// null buffer argument should fail with TypeError.
			try {
				bytesRead = await file.read(null)
			} catch(err) {
				if (err.name !== 'TypeError') {
					throw 'unexpected error: ' + err;
				}
			}

			// undefined buffer argument should fail with TypeError.
			try {
				bytesRead = await file.read(undefined)
			} catch (err) {
				if (err.name !== 'TypeError') {
					throw 'unexpected error: ' + err;
				}
			}

			// Invalid typed array argument should fail with TypeError.
			try {
				bytesRead = await file.read(new Int32Array(5))
			} catch (err) {
				if (err.name !== 'TypeError') {
					throw 'unexpected error: ' + err;
				}
			}

			// ArrayBuffer argument should fail with TypeError.
			try {
				bytesRead = await file.read(new ArrayBuffer(5))
			} catch (err) {
				if (err.name !== 'TypeError') {
					throw 'unexpected error: ' + err;
				}
			}
		`, testFilePath)))

		assert.NoError(t, err)
	})

	t.Run("read on closed file should fail", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		testFilePath := fsext.FilePathSeparator + "bonjour.txt"
		fs := newTestFs(t, func(fs afero.Fs) error {
			return afero.WriteFile(fs, testFilePath, []byte("hello"), 0o644)
		})
		runtime.VU.InitEnvField.FileSystems["file"] = fs

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(fmt.Sprintf(`
			const file = await fs.open(%q)
			await file.close()

			try {
				const bytesRead = await file.read(new Uint8Array(3))
				throw 'unexpected promise resolution with result: ' + bytesRead;
			} catch (err) {
				if (err.name !== 'BadResourceError') {
					throw 'unexpected error: ' + err
				}
			}
		`, testFilePath)))

		assert.NoError(t, err)
	})

	t.Run("seek with invalid arguments should fail", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		testFilePath := fsext.FilePathSeparator + "bonjour.txt"
		fs := newTestFs(t, func(fs afero.Fs) error {
			return afero.WriteFile(fs, testFilePath, []byte("hello"), 0o644)
		})
		runtime.VU.InitEnvField.FileSystems["file"] = fs

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(fmt.Sprintf(`
			const file = await fs.open(%q)

			let newOffset

			// null offset should fail with TypeError.
			try {
				newOffset = await file.seek(null)
				throw "file.seek(null) promise unexpectedly resolved with result: " + newOffset
			} catch (err) {
				if (err.name !== 'TypeError') {
					throw "file.seek(null) rejected with unexpected error: " + err
				}
			}

			// undefined offset should fail with TypeError.
			try {
				newOffset = await file.seek(undefined)
				throw "file.seek(undefined) promise unexpectedly promise resolved with result: " + newOffset
			} catch (err) {
				if (err.name !== 'TypeError') {
					throw "file.seek(undefined) rejected with unexpected error: " + err
				}
			}

			// Invalid type offset should fail with TypeError.
			try {
				newOffset = await file.seek('abc')
				throw "file.seek('abc') promise unexpectedly resolved with result: " + newOffset 
			} catch (err) {
				if (err.name !== 'TypeError') {
					throw "file.seek('1') rejected with unexpected error: " + err
				}
			}

			// Negative offset should fail with TypeError.
			try {
				newOffset = await file.seek(-1)
				throw "file.seek(-1) promise unexpectedly resolved with result: " + newOffset
			} catch (err) {
				if (err.name !== 'TypeError') {
					throw "file.seek(-1) rejected with unexpected error: " + err
				}
			}

			// Invalid type whence should fail with TypeError.
			try {
				newOffset = await file.seek(1, 'abc')
				throw "file.seek(1, 'abc') promise unexpectedly resolved with result: " + newOffset
			} catch (err) {
				if (err.name !== 'TypeError') {
					throw "file.seek(1, 'abc') rejected with unexpected error: " + err
				}
			}

			// Invalid whence should fail with TypeError.
			try {
				newOffset = await file.seek(1, -1)
				throw "file.seek(1, -1) promise unexpectedly resolved with result: " + newOffset 
			} catch (err) {
				if (err.name !== 'TypeError') {
					throw "file.seek(1, -1) rejected with unexpected error: " + err
				}
			}
		`, testFilePath)))

		assert.NoError(t, err)
	})

	t.Run("seek on closed file should fail", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		testFilePath := fsext.FilePathSeparator + "bonjour.txt"
		fs := newTestFs(t, func(fs afero.Fs) error {
			return afero.WriteFile(fs, testFilePath, []byte("hello"), 0o644)
		})
		runtime.VU.InitEnvField.FileSystems["file"] = fs

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(fmt.Sprintf(`
			const file = await fs.open(%q)
			await file.close()

			try {
				const newOffset = await file.seek(0, fs.SeekMode.Start)
				throw 'unexpected promise resolution with result: ' + newOffset;
			} catch (err) {
				if (err.name !== 'BadResourceError') {
					throw 'unexpected error: ' + err
				}
			}
		`, testFilePath)))

		assert.NoError(t, err)
	})

	t.Run("double close should fail", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		testFilePath := fsext.FilePathSeparator + "bonjour.txt"
		fs := newTestFs(t, func(fs afero.Fs) error {
			return afero.WriteFile(fs, testFilePath, []byte("hello"), 0o644)
		})
		runtime.VU.InitEnvField.FileSystems["file"] = fs

		_, err = runtime.RunOnEventLoop(wrapInAsyncLambda(fmt.Sprintf(`
			const file = await fs.open(%q)

			await file.close()

			try {
				await file.close()
				throw 'unexpected promise resolution';
			} catch (err) {
				if (err.name !== 'BadResourceError') {
					throw 'unexpected error: ' + err
				}
			}
		`, testFilePath)))

		assert.NoError(t, err)
	})
}

func TestOpenImpl(t *testing.T) {
	t.Parallel()

	t.Run("should panic if the file system is not available", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)
		delete(runtime.VU.InitEnvField.FileSystems, "file")

		mi := &ModuleInstance{
			vu:    runtime.VU,
			cache: &cache{},
		}

		assert.Panics(t, func() {
			//nolint:errcheck,gosec
			mi.openImpl("bonjour.txt")
		})
	})

	t.Run("should return an error if the file does not exist", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		mi := &ModuleInstance{
			vu:    runtime.VU,
			cache: &cache{},
		}

		_, err = mi.openImpl("bonjour.txt")
		assert.Error(t, err)
		var fsError *fsError
		assert.ErrorAs(t, err, &fsError)
		assert.Equal(t, NotFoundError, fsError.kind)
	})

	t.Run("should return an error if the path is a directory", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		fs := newTestFs(t, func(fs afero.Fs) error {
			return fs.Mkdir("/dir", 0o644)
		})
		runtime.VU.InitEnvField.FileSystems["file"] = fs

		mi := &ModuleInstance{
			vu:    runtime.VU,
			cache: &cache{},
		}

		_, err = mi.openImpl("/dir")
		assert.Error(t, err)
		var fsError *fsError
		assert.ErrorAs(t, err, &fsError)
		assert.Equal(t, InvalidResourceError, fsError.kind)
	})

	t.Run("path is resolved relative to the entrypoint script", func(t *testing.T) {
		t.Parallel()

		runtime, err := newConfiguredRuntime(t)
		require.NoError(t, err)

		fs := newTestFs(t, func(fs afero.Fs) error {
			return afero.WriteFile(fs, "/bonjour.txt", []byte("Bonjour, le monde"), 0o644)
		})
		runtime.VU.InitEnvField.FileSystems["file"] = fs
		runtime.VU.InitEnvField.CWD = &url.URL{Scheme: "file", Path: "/dir"}

		mi := &ModuleInstance{
			vu:    runtime.VU,
			cache: &cache{},
		}

		_, err = mi.openImpl("../bonjour.txt")
		assert.NoError(t, err)
	})
}

const initGlobals = `
	globalThis.fs = require("k6/experimental/fs");
`

func newConfiguredRuntime(t testing.TB) (*modulestest.Runtime, error) {
	runtime := modulestest.NewRuntime(t)

	err := runtime.SetupModuleSystem(map[string]interface{}{"k6/experimental/fs": New()}, nil, compiler.New(runtime.VU.InitEnv().Logger))
	if err != nil {
		return nil, err
	}

	// Set up the VU environment with an in-memory filesystem and a CWD of "/".
	runtime.VU.InitEnvField.FileSystems = map[string]fsext.Fs{
		"file": fsext.NewMemMapFs(),
	}
	runtime.VU.InitEnvField.CWD = &url.URL{Scheme: "file"}

	// Ensure the `fs` module is available in the VU's runtime.
	_, err = runtime.VU.Runtime().RunString(initGlobals)

	return runtime, err
}

// newTestFs is a helper function that creates a new in-memory file system and calls the provided
// function with it. The provided function is expected to use the file system to create files and
// directories.
func newTestFs(t *testing.T, fn func(fs afero.Fs) error) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	err := fn(fs)
	if err != nil {
		t.Fatal(err)
	}

	return fs
}

// wrapInAsyncLambda is a helper function that wraps the provided input in an async lambda. This
// makes the use of `await` statements in the input possible.
func wrapInAsyncLambda(input string) string {
	// This makes it possible to use `await` freely on the "top" level
	return "(async () => {\n " + input + "\n })()"
}
