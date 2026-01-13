<script>
  import { onMount } from 'svelte';
  import { GetAllVersions } from '../wailsjs/go/main/App.js';

  export let antiTamperStatus;
  export let zdpServiceStatus;
  export let dlpSdkVersion;

  let otherVersions = {
    zdp: 'Loading...',
    zcc: 'Loading...',
    zep: 'Loading...'
  };

  onMount(async () => {
    try {
      const versions = await GetAllVersions();
      otherVersions = versions;
    } catch (error) {
      console.error("Failed to get all versions:", error);
      otherVersions = {
        zdp: 'Error',
        zcc: 'Error',
        zep: 'Error'
      };
    }
  });

  $: antiTamperTextColor = getAntiTamperTextColor(antiTamperStatus);
  $: zdpServiceTextColor = getZdpServiceTextColor(zdpServiceStatus);

  function getAntiTamperTextColor(status) {
    switch (status) {
      case 'Enabled':
        return 'green-text';
      case 'Disabled':
        return 'red-text';
      default:
        return 'grey-text'; // For "Unknown" or other states
    }
  }

  function getZdpServiceTextColor(status) {
    switch (status) {
      case 'Running':
        return 'green-text';
      case 'Starting': // Assuming these states might exist
      case 'Stopping':
        return 'yellow-text';
      case 'Stopped':
        return 'red-text';
      default:
        return 'grey-text'; // For "Unknown" or other states
    }
  }
</script>

<footer>
  <div class="status-bar">
    <span>Anti-tampering: <span class="{antiTamperTextColor}">{antiTamperStatus}</span></span>
    <span>ZDP Service: <span class="{zdpServiceTextColor}">{zdpServiceStatus}</span></span>
    <span>DLP SDK Version: <span class="grey-text">{dlpSdkVersion}</span></span>
    <span>ZDP: <span class="grey-text">{otherVersions.zdp}</span></span>
    <span>ZCC: <span class="grey-text">{otherVersions.zcc}</span></span>
    <span>ZEP: <span class="grey-text">{otherVersions.zep}</span></span>
  </div>
</footer>

<style>
  footer {
    position: fixed;
    bottom: 0;
    left: 0;
    width: 100%;
    background-color: #1b2636;
    padding: 0.5em;
    border-top: 1px solid #333;
    z-index: 1000; /* Ensure footer is above other content */
    color: white; /* Ensure text is visible on dark background */
  }

  .status-bar {
    display: flex;
    justify-content: space-around;
    font-size: 0.8em;
    align-items: center; /* Align items vertically in the middle */
  }

  .green-text {
    color: #4CAF50; /* Green */
  }

  .yellow-text {
    color: #FFC107; /* Yellow */
  }

  .red-text {
    color: #F44336; /* Red */
  }

  .grey-text {
    color: #9E9E9E; /* Grey */
  }
</style>
