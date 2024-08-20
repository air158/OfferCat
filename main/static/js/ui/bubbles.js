const bubblesContainer = document.getElementById('bubbles-container');
const upperSection = document.getElementById('upper-section');
const submitButton = document.getElementById('submit-button');
const lowerSection = document.getElementById('lower-section');
const responseBox = document.getElementById('response-box');
const responseContent = document.getElementById('response-content');
const closeButton = document.getElementById('close-button');

const interviewerQuesiton = document.getElementById('interviewer-quesiton');

let currentBubble = null;
let isSelecting = false;
let shouldScrollToBottom = true;

function createBubble() {
    const bubble = document.createElement('div');
    bubble.className = 'bubble';
    
    bubble.addEventListener('mousedown', (e) => {
        isSelecting = true;
        toggleSelection(bubble);
    });

    bubble.addEventListener('mouseover', () => {
        if (isSelecting) {
            toggleSelection(bubble);
        }
    });

    bubblesContainer.appendChild(bubble);
    if (shouldScrollToBottom) {
        scrollToBottom();
    }

    var display = window.getComputedStyle(interviewerQuesiton).display;
    if (display === 'none') {
        console.log('The element is hidden (display: none).');
        interviewerQuesiton.style.display = 'block';
        interviewerQuesiton.className = 'bubble';
        interviewerQuesiton.addEventListener('mousedown', (e) => {
            isSelecting = true;
            toggleSelection(interviewerQuesiton);
        });
    
        interviewerQuesiton.addEventListener('mouseover', () => {
            if (isSelecting) {
                toggleSelection(interviewerQuesiton);
            }
        });
    } else {
        console.log('The element is visible.');
    }
    return bubble;
}

function updateBubble(bubble, text) {
    bubble.textContent = text;
    if (shouldScrollToBottom) {
        scrollToBottom();
    }
}

function toggleSelection(bubble) {
    bubble.classList.toggle('selected');
    // updateSubmitButtonVisibility(bubble);
}

function updateSubmitButtonVisibility(bubble) {
    const selectedBubbles = document.querySelectorAll('.bubble.selected');
    if (selectedBubbles.length > 0) {
        submitButton.style.removeProperty('display');
        // submitButton.style.display = 'block';
        // positionSubmitButton(selectedBubbles[selectedBubbles.length - 1]);
    } else {
        submitButton.style.display = 'none';
    }
}

function positionSubmitButton(bubble) {
    const bubbleRect = bubble.getBoundingClientRect();
    const containerRect = bubblesContainer.getBoundingClientRect();
    
    // submitButton.style.top = `${bubbleRect.top - containerRect.top + (bubbleRect.height - submitButton.offsetHeight) / 2}px`;

    submitButton.style.top = `${bubbleRect.top - containerRect.top + (bubbleRect.height / 2) - 3}px`;
    submitButton.style.left = `${bubbleRect.left - containerRect.left - submitButton.offsetWidth - 10}px`; // 10px作为间距
    submitButton.style.display = 'block';
}

function scrollToBottom() {
    upperSection.scrollTop = upperSection.scrollHeight;
}

upperSection.addEventListener('scroll', () => {
    const isAtBottom = upperSection.scrollHeight - upperSection.scrollTop === upperSection.clientHeight;
    shouldScrollToBottom = isAtBottom;
});

document.addEventListener('mouseup', () => {
    isSelecting = false;
});

submitButton.addEventListener('click', () => {
    const selectedBubbles = document.querySelectorAll('.bubble.selected');
    const selectedText = Array.from(selectedBubbles).map(bubble => bubble.textContent).join(' ');
    AIResponse(selectedText);
});



closeButton.addEventListener('click', closeResponseBox);
lowerSection.addEventListener('click', (e) => {
    if (e.target === lowerSection) {
        closeResponseBox();
    }
});

function closeResponseBox() {
    lowerSection.style.display = 'none';
    responseContent.innerHTML = '';
}

function clearAllBubbles() {
    bubblesContainer.innerHTML = '';
    // bubblesContainer.appendChild(submitButton);
    // updateSubmitButtonVisibility();
}

function setInterviewerQuesiton(quesiton) {
    interviewerQuesiton.innerText = quesiton
}