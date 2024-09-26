const SERVER_URL = window.location.hostname + ':8080';

const toastifyCallback = (body) => {
    Toastify({
        text: body,
        duration: 3000,
        newWindow: true,
        escapeMarkup: false
    }).showToast();
}

function noticeMe(callback)  {
    const ws = new WebSocket(`ws://${SERVER_URL}/ws`);

    ws.onmessage = function(e) {
        const messages = e.data.split('\n');
        for (let i = 0; i < messages.length; i++) {
            if(callback && callback instanceof Function) {
                callback(messages[i]);
            } else {
                toastifyCallback(messages[i]);
            }
        }
    };

    ws.onclose = function() {
        setTimeout(function() {
            noticeMe(callback);
        }, 3000);
    };

    ws.onerror = function() {
        ws.close();
    };
}

export default noticeMe;
