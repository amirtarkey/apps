<script>
    import { onMount, onDestroy } from 'svelte';
    import {
        IsZdpServiceRunning,
        GetDetailsHttpsCmd,
        GetDetailsHttpCmd,
        EnableAntiTampering,
        DisableAntiTampering,
        GetAntiTamperingStatus,
        DeobfuscateOotbSettings,
        DeobfuscateZdpModes,
        IsOotbSettingsObfuscated,
        IsZdpModesObfuscated
    } from '../wailsjs/go/main/App.js';
    import { ClipboardSetText } from '../wailsjs/runtime/runtime.js';
    import Footer from './Footer.svelte';
  
    let antiTamperStatus = '';
    let resultText = '';
    let endpointDetails = null;
    let copyButtonText = 'Copy';
    let zdpServiceStatus = '';
    let isOotbSettingsObfuscated = false;
    let isZdpModesObfuscated = false;
  
    let timeoutId = null;
    let intervalId = null;
  
    function clearResultText() {
      resultText = '';
      endpointDetails = null;
    }
  
    function autoClearResultText(timeout = 10000) {
      console.log(`autoClearResultText called with timeout: ${timeout}`);
      if (timeoutId) {
        clearTimeout(timeoutId);
      }
      timeoutId = setTimeout(() => {
        console.log('autoClearResultText: clearing result text');
        resultText = '';
        endpointDetails = null;
      }, timeout);
    }
  
    async function checkStatuses() {
      try {
        antiTamperStatus = await GetAntiTamperingStatus();
  	    let isRunning = await IsZdpServiceRunning();
  	    zdpServiceStatus = isRunning ? 'Running' : 'Stopped';
        isOotbSettingsObfuscated = await IsOotbSettingsObfuscated();
        isZdpModesObfuscated = await IsZdpModesObfuscated();
      } catch (error) {
        resultText = `Error: ${error}`;
        autoClearResultText(1000);
      }
    }
    
    async function handleToggleAntiTampering(event) {
      const isEnabled = event.target.checked;
      clearResultText();
      try {
        if (isEnabled) {
          resultText = 'Enabling anti-tampering...';
          await EnableAntiTampering();
          resultText = 'Anti-tampering enabled successfully.';
        } else {
          resultText = 'Disabling anti-tampering...';
          await DisableAntiTampering();
          resultText = 'Anti-tampering disabled successfully.';
        }
        checkStatuses();
      } catch (error) {
        resultText = `Error: ${error}`;
      }
      autoClearResultText(1000);
    }
  
    async function handleGetEndpointDetails() {
      clearResultText();
      resultText = 'Getting endpoint details...';
      endpointDetails = null;
      console.log('handleGetEndpointDetails: start');
  
      try {
        const isRunning = await IsZdpServiceRunning();
        if (!isRunning) {
          resultText = 'ZDP service is not running.';
          autoClearResultText(10000);
          console.log('handleGetEndpointDetails: ZDP service not running');
          return;
        }
  
        try {
          console.log('handleGetEndpointDetails: trying HTTPS');
          const details = await GetDetailsHttpsCmd();
          endpointDetails = JSON.parse(details);
          resultText = 'Endpoint details retrieved successfully via HTTPS.';
          console.log('handleGetEndpointDetails: HTTPS success');
          autoClearResultText(10000);
        } catch (httpsError) {
          console.log(`handleGetEndpointDetails: HTTPS error: ${httpsError}`);
          resultText = `HTTPS attempt failed: ${httpsError}. Trying HTTP...`;
          try {
            console.log('handleGetEndpointDetails: trying HTTP');
            const details = await GetDetailsHttpCmd();
            endpointDetails = JSON.parse(details);
            resultText = 'Endpoint details retrieved successfully via HTTP.';
            console.log('handleGetEndpointDetails: HTTP success');
            autoClearResultText(10000);
          } catch (httpError) {
            console.log(`handleGetEndpointDetails: HTTP error: ${httpError}`);
            resultText = `HTTP attempt also failed: ${httpError}`;
            autoClearResultText(10000);
          }
        }
      } catch (error) {
        console.log(`handleGetEndpointDetails: unexpected error: ${error}`);
        resultText = `An unexpected error occurred: ${error}`;
        autoClearResultText(10000);
      }
    }
  
    async function handleDeobfuscateOotbSettings() {
      clearResultText();
      console.log('handleDeobfuscateOotbSettings: start');
      try {
        resultText = 'De-obfuscating ootb-settings...';
        const result = await DeobfuscateOotbSettings();
        resultText = result;
        console.log(`handleDeobfuscateOotbSettings: success: ${result}`);
        checkStatuses();
      } catch (error) {
        resultText = `Error: ${error}`;
        console.log(`handleDeobfuscateOotbSettings: error: ${error}`);
      }
      autoClearResultText(5000);
    }
  
    async function handleDeobfuscateZdpModes() {
      clearResultText();
      console.log('handleDeobfuscateZdpModes: start');
      try {
        resultText = 'De-obfuscating zdp-modes...';
        const result = await DeobfuscateZdpModes();
        resultText = result;
        console.log(`handleDeobfuscateZdpModes: success: ${result}`);
        checkStatuses();
      } catch (error) {
        resultText = `Error: ${error}`;
        console.log(`handleDeobfuscateZdpModes: error: ${error}`);
      }
      autoClearResultText(5000);
    }
  
    function copyEndpointDetails() {
      if (endpointDetails) {
        ClipboardSetText(JSON.stringify(endpointDetails, null, 2));
        copyButtonText = 'Copied!';
        setTimeout(() => {
          copyButtonText = 'Copy';
        }, 2000);
      }
    }
  
    onMount(() => {
      checkStatuses();
      intervalId = setInterval(checkStatuses, 5000); // Check every 5 seconds
    });

    onDestroy(() => {
      if (intervalId) {
        clearInterval(intervalId);
      }
    });
  </script>
  
  <main>
    <h1>ZDP Tool</h1>
    <div class="buttons">
          <div class="toggle-container">
            <label for="anti-tamper-toggle">Anti-tampering</label>
            <label class="switch">
              <input
                id="anti-tamper-toggle"
                type="checkbox"
                checked={antiTamperStatus === 'Enabled'}
                on:change={handleToggleAntiTampering}
              />
              <span class="slider round"></span>
            </label>
          </div>
          <button on:click={handleGetEndpointDetails}>Get Endpoint Details</button>
              <div class="toggle-container">
                <label for="ootb-settings-toggle">De-obfuscate ootb-settings</label>
                <label class="switch">
                  <input
                    id="ootb-settings-toggle"
                    type="checkbox"
                    checked={!isOotbSettingsObfuscated}
                    disabled={!isOotbSettingsObfuscated}
                    on:change={handleDeobfuscateOotbSettings}
                  />
                  <span class="slider round"></span>
                </label>
              </div>
              <div class="toggle-container">
                <label for="zdp-modes-toggle">De-obfuscate zdp-modes</label>
                <label class="switch">
                  <input
                    id="zdp-modes-toggle"
                    type="checkbox"
                    checked={!isZdpModesObfuscated}
                    disabled={!isZdpModesObfuscated}
                    on:change={handleDeobfuscateZdpModes}
                  />
                  <span class="slider round"></span>
                </label>
              </div>
            </div>    <div class="output">
      <div class="result">
        {resultText}
      </div>
      {#if endpointDetails}
        <div class="endpoint-details">
          <div class="header">
            <h2>Endpoint Details</h2>
            <button on:click={copyEndpointDetails}>{copyButtonText}</button>
          </div>
          <pre>{JSON.stringify(endpointDetails, null, 2)}</pre>
        </div>
      {/if}
    </div>
  </main>
  
  <Footer {antiTamperStatus} {zdpServiceStatus} />
  
  <style>
    main {
      padding: 1em;
      display: flex;
      flex-direction: column;
      align-items: center;
      height: 100vh;
    }
  
    h1 {
      color: #ff3e00;
      text-transform: uppercase;
      font-size: 2em;
      font-weight: 100;
    }
  
    .buttons {
      display: flex;
      flex-direction: column;
      gap: 0.5em;
      width: 100%;
    }
  
    .output {
      margin-top: 1em;
      width: 100%;
    }
  
    .result {
      margin-top: 1em;
    }
  
    .endpoint-details {
      margin-top: 1em;
      text-align: left;
    }
  
      .endpoint-details .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
      }
    
      .toggle-container {
        display: flex;
        justify-content: space-between;
        align-items: center;
        width: 100%;
        margin-bottom: 0.5em;
      }
    
      .switch {
        position: relative;
        display: inline-block;
        width: 50px;
        height: 24px;
      }
    
      .switch input { 
        opacity: 0;
        width: 0;
        height: 0;
      }
    
      .slider {
        position: absolute;
        cursor: pointer;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background-color: #ccc;
        -webkit-transition: .4s;
        transition: .4s;
      }
    
      .slider:before {
        position: absolute;
        content: "";
        height: 16px;
        width: 16px;
        left: 4px;
        bottom: 4px;
        background-color: white;
        -webkit-transition: .4s;
        transition: .4s;
      }
    
      input:checked + .slider {
        background-color: #2196F3;
      }
    
      input:focus + .slider {
        box-shadow: 0 0 1px #2196F3;
      }
    
      input:checked + .slider:before {
        -webkit-transform: translateX(26px);
        -ms-transform: translateX(26px);
        transform: translateX(26px);
      }
    
      /* Rounded sliders */
      .slider.round {
        border-radius: 34px;
      }
    
      .slider.round:before {
        border-radius: 50%;
      }
    </style>
