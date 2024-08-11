// JavaScript to toggle the settings container and button visibility
document.getElementById('settings-button').addEventListener('click', function() {
    var settingsContainer = document.getElementById('settings-container');
    var settingsButton = document.getElementById('settings-button');
    settingsContainer.style.display = 'block';
    settingsButton.style.display = 'none';
    start();
});

document.getElementById('close-settings-button').addEventListener('click', function() {
    var settingsContainer = document.getElementById('settings-container');
    var settingsButton = document.getElementById('settings-button');
    settingsContainer.style.display = 'none';
    settingsButton.style.display = 'block';
});

document.addEventListener('click', function(event) {
    var settingsContainer = document.getElementById('settings-container');
    var settingsButton = document.getElementById('settings-button');
    if (!settingsContainer.contains(event.target) && !settingsButton.contains(event.target)) {
        settingsContainer.style.display = 'none';
        settingsButton.style.removeProperty('display');
        // settingsButton.style.display = 'block';
    }
});

// 添加提示界面关闭逻辑
document.getElementById('tooltip-close').addEventListener('click', function() {
    tooltip.style.display = 'none';
    overlay.style.display = 'none';
});

overlay.addEventListener('click', function() {
    tooltip.style.display = 'none';
    overlay.style.display = 'none';
});

// 点击提示界面外部关闭提示界面
document.addEventListener('click', function(event) {
    if (!tooltip.contains(event.target)) {
        tooltip.style.display = 'none';
        overlay.style.display = 'none';
    }
});