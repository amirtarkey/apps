<script>
  import { StandaloneClassifier, SelectFile, ReadFileContent } from '../wailsjs/go/main/App';
  import { ClipboardSetText } from '../wailsjs/runtime/runtime';

  let filePath = '';
  let configOption = 'default';
  let customConfigPath = '';
  let useOcr = false;
  let useText = false;
  let resultText = '';
  let parsedOutput = null;
  let executedCommand = '';
  let ocrTextOutput = '';
  let extractedTextOutput = '';
  
  let showExecutedCommand = false;
  let showRawOutput = false;
  let showOcrTextOutput = false;
  let showExtractedTextOutput = false;

  function parseOutput(output) {
    const result = {};
    output = output.replace(/\r/g, "");

    function compareVersions(v1, v2) {
        const parts1 = v1.split('.').map(Number);
        const parts2 = v2.split('.').map(Number);
        const len = Math.max(parts1.length, parts2.length);
        for (let i = 0; i < len; i++) {
            const p1 = parts1[i] || 0;
            const p2 = parts2[i] || 0;
            if (p1 > p2) return 1;
            if (p1 < p2) return -1;
        }
        return 0;
    }

    let match = output.match(/DLP SDK version: (.*)/);
    if (match) result["DLP SDK version"] = match[1].trim();

    const dlpSdkVersion = result["DLP SDK version"] || "0.0.0";
    const withSubDictionaries = compareVersions(dlpSdkVersion, "4.0.0") > 0;

    match = output.match(/File path: (.*)/);
    if (match) result["File path"] = match[1].trim();

    match = output.match(/Result: (.*)/);
    if (match) result["Result"] = match[1].trim();

    match = output.match(/File types: (.*)/);
    if (match) result["File types"] = match[1].trim();

    match = output.match(/Supported file type: (.*)/);
    if (match) result["Supported file type"] = match[1].trim();

    match = output.match(/Doc category: (\d+)/);
    if (match) result["Doc category"] = match[1].trim();
    
    match = output.match(/Doc subcategory: (\d+)/);
    if (match) result["Doc subcategory"] = match[1].trim();

    match = output.match(/Scan time ms: (\d+)/);
    if (match) result["Scan time ms"] = match[1].trim();

    match = output.match(/Engines:\s*([\s\S]*?)Dictionaries:/);
    if (match) {
        const enginesStr = match[1].trim();
        result["Engines"] = {};
        const engineParts = enginesStr.replace(/]\s*\[/g, '|||').split('|||');
        for (const part of engineParts) {
            const engineMatch = part.match(/(\d+):\s*(.*)/);
            if (engineMatch) {
                result["Engines"][engineMatch[1].trim()] = engineMatch[2].trim();
            }
        }
    }

    match = output.match(/Dictionaries:\s*([\s\S]*?)(?=\nEDM match:|$)/);
    if (match) {
        const dictionariesStr = match[1].trim();
        result["Dictionaries"] = {};
        if (withSubDictionaries) {
            result["Sub Dictionaries"] = {};
        }

        const dictRegex = /(\d+)(?:,(\d+))?:\s*([^;]+)/g;
        let dictMatch;
        while ((dictMatch = dictRegex.exec(dictionariesStr)) !== null) {
            const [, dictId, subDictId, dictName] = dictMatch;
            result["Dictionaries"][dictId.trim()] = dictName.trim();
            if (withSubDictionaries && subDictId) {
                result["Sub Dictionaries"][dictId.trim()] = subDictId.trim();
            }
        }
    }

    match = output.match(/EDM match: (.*)/);
    if (match) {
        try {
            result["EDM match"] = JSON.parse(match[1]);
        } catch (e) {
            console.error("Failed to parse EDM match JSON:", e);
        }
    }

    return result;
  }

  async function classify() {
    if (!filePath) {
      resultText = 'Please select a file to classify.';
      parsedOutput = null;
      executedCommand = '';
      ocrTextOutput = '';
      extractedTextOutput = '';
      return;
    }
    resultText = 'Classifying...';
    parsedOutput = null;
    executedCommand = '';
    ocrTextOutput = '';
    extractedTextOutput = '';

    try {
      const result = await StandaloneClassifier(filePath, configOption, customConfigPath, useOcr, useText);
      executedCommand = result.command;
      resultText = result.output;
      parsedOutput = parseOutput(result.output);

      if (result.ocrTextPath) {
        try {
          ocrTextOutput = await ReadFileContent(result.ocrTextPath);
        } catch (e) {
          ocrTextOutput = `Error reading OCR text file: ${e}`;
        }
      }

      if (result.extractedTextPath) {
        try {
          extractedTextOutput = await ReadFileContent(result.extractedTextPath);
        } catch (e) {
          extractedTextOutput = `Error reading extracted text file: ${e}`;
        }
      }

    } catch (error) {
      resultText = `Error: ${error}`;
      parsedOutput = null;
      executedCommand = '';
    }
  }

  async function selectFile() {
    try {
      const selectedFile = await SelectFile();
      if (selectedFile) {
        filePath = selectedFile;
      }
    } catch (error) {
      console.error(error);
    }
  }

  async function selectConfigFile() {
    try {
      const selectedFile = await SelectFile();
      if (selectedFile) {
        customConfigPath = selectedFile;
      }
    } catch (error) {
      console.error(error);
    }
  }

  function copyOcrTextOutput() {
    if (ocrTextOutput) {
      ClipboardSetText(ocrTextOutput);
    }
  }

  function copyExtractedTextOutput() {
    if (extractedTextOutput) {
      ClipboardSetText(extractedTextOutput);
    }
  }
</script>

<main>
  <div class="container">
    <h2>Standalone Classifier</h2>
    <div class="form-group">
      <label for="file-path">File for Classification:</label>
      <input id="file-path" type="text" bind:value={filePath} readonly />
      <button on:click={selectFile}>Select File</button>
    </div>

    <div class="form-group">
      <label>Config File:</label>
      <div>
        <input type="radio" id="default-config" value="default" bind:group={configOption}>
        <label for="default-config">Default</label>
      </div>
      <div>
        <input type="radio" id="last-modified-config" value="last_modified" bind:group={configOption}>
        <label for="last-modified-config">Last Modified</label>
      </div>
      <div>
        <input type="radio" id="custom-config" value="custom" bind:group={configOption}>
        <label for="custom-config">Custom</label>
      </div>
    </div>

    {#if configOption === 'custom'}
    <div class="form-group">
      <label for="custom-config-path">Custom Config Path:</label>
      <input id="custom-config-path" type="text" bind:value={customConfigPath} readonly />
      <button on:click={selectConfigFile}>Select Config File</button>
    </div>
    {/if}

    <div class="form-group">
      <label for="ocr">Use OCR:</label>
      <input id="ocr" type="checkbox" bind:checked={useOcr} />
    </div>

    <div class="form-group">
      <label for="text">Use Text:</label>
      <input id="text" type="checkbox" bind:checked={useText} disabled={!useOcr} />
    </div>

    <button on:click={classify}>Classify</button>
    <div class="output">
      {#if executedCommand}
        <div class="result">
          <div class="output-header">
            <h3>Executed Command</h3>
            <button on:click={() => showExecutedCommand = !showExecutedCommand}>
              {showExecutedCommand ? 'Collapse' : 'Expand'}
            </button>
          </div>
          {#if showExecutedCommand}
            <pre>{executedCommand}</pre>
          {/if}
        </div>
      {/if}

      {#if parsedOutput}
        <div class="parsed-output">
          <h3>Classification Result</h3>
          {#each Object.entries(parsedOutput) as [key, value]}
            <div class="kv-pair">
              <span class="key">{key}:</span>
              {#if typeof value === 'object' && value !== null}
                <div class="nested">
                  {#each Object.entries(value) as [nestedKey, nestedValue]}
                    <div class="kv-pair">
                      <span class="key">{nestedKey}:</span>
                      <span>{nestedValue}</span>
                    </div>
                  {/each}
                </div>
              {:else}
                <span>{value}</span>
              {/if}
            </div>
          {/each}
        </div>
      {/if}

      {#if resultText}
        <div class="result">
          <div class="output-header">
            <h3>Raw Output</h3>
            <button on:click={() => showRawOutput = !showRawOutput}>
              {showRawOutput ? 'Collapse' : 'Expand'}
            </button>
          </div>
          {#if showRawOutput}
            <pre>{resultText}</pre>
          {/if}
        </div>
      {/if}

      {#if ocrTextOutput}
        <div class="text-output">
          <div class="output-header">
            <h3>OCR Text Output</h3>
            <button on:click={() => showOcrTextOutput = !showOcrTextOutput}>
              {showOcrTextOutput ? 'Collapse' : 'Expand'}
            </button>
            <button on:click={copyOcrTextOutput}>Copy</button>
          </div>
          {#if showOcrTextOutput}
            <pre>{ocrTextOutput}</pre>
          {/if}
        </div>
      {/if}

      {#if extractedTextOutput}
        <div class="text-output">
          <div class="output-header">
            <h3>Extracted Text Output</h3>
            <button on:click={() => showExtractedTextOutput = !showExtractedTextOutput}>
              {showExtractedTextOutput ? 'Collapse' : 'Expand'}
            </button>
            <button on:click={copyExtractedTextOutput}>Copy</button>
          </div>
          {#if showExtractedTextOutput}
            <pre>{extractedTextOutput}</pre>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</main>

<style>
  .output {
    margin-top: 1em;
    width: 100%;
    text-align: left;
  }
  .parsed-output, .text-output {
    border: 1px solid #ccc;
    padding: 1em;
    border-radius: 5px;
    margin-bottom: 1em;
  }
  .kv-pair {
    display: flex;
    margin-bottom: 0.5em;
  }
  .key {
    font-weight: bold;
    margin-right: 0.5em;
  }
  .nested {
    margin-left: 1em;
    border-left: 1px solid #eee;
    padding-left: 1em;
  }
  pre {
    white-space: pre-wrap;
    word-wrap: break-word;
  }
  .output-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5em;
  }
  .output-header h3 {
    margin: 0;
  }
  .output-header button {
    margin-left: 0.5em;
  }
</style>