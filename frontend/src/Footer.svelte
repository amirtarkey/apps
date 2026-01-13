<script>
  export let antiTamperStatus;
  export let zdpServiceStatus;

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

