<script>
  export let antiTamperStatus;
  export let zdpServiceStatus;

  $: antiTamperDotColor = getAntiTamperDotColor(antiTamperStatus);
  $: zdpServiceDotColor = getZdpServiceDotColor(zdpServiceStatus);

  function getAntiTamperDotColor(status) {
    switch (status) {
      case 'Enabled':
        return 'green';
      case 'Disabled':
        return 'red';
      default:
        return 'grey'; // For "Unknown" or other states
    }
  }

  function getZdpServiceDotColor(status) {
    switch (status) {
      case 'Running':
        return 'green';
      case 'Starting': // Assuming these states might exist
      case 'Stopping':
        return 'yellow';
      case 'Stopped':
        return 'red';
      default:
        return 'grey'; // For "Unknown" or other states
    }
  }
</script>

<footer>
  <div class="status-bar">
        <span>Anti-tampering: <span class="status-dot {antiTamperDotColor}"></span> {antiTamperStatus}</span>
        <span>ZDP Service: <span class="status-dot {zdpServiceDotColor}"></span> {zdpServiceStatus}</span>
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

  .status-dot {
    height: 8px;
    width: 8px;
    background-color: grey; /* Default color */
    border-radius: 50%;
    display: inline-block;
    margin-right: 5px; /* Space between dot and text */
  }

  .status-dot.green {
    background-color: #4CAF50; /* Green */
  }

  .status-dot.yellow {
    background-color: #FFC107; /* Yellow */
  }

  .status-dot.red {
    background-color: #F44336; /* Red */
  }

  .status-dot.grey {
    background-color: #9E9E9E; /* Grey */
  }
</style>
