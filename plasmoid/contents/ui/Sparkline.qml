import QtQuick

// A small line graph of 0 - 100 values. New samples appear on the right;
// "capacity" fixes the spacing so the graph scrolls instead of stretching.
Canvas {
    id: canvas

    // [{ values: [..], color: "..." }, ...]
    property var series: []
    property int capacity: 120
    property real lineWidth: 1.5
    property bool fill: true
    // Horizontal guide lines, in percent. Empty for none.
    property var guides: []
    property color guideColor: "gray"

    onSeriesChanged: requestPaint()
    onCapacityChanged: requestPaint()
    onWidthChanged: requestPaint()
    onHeightChanged: requestPaint()

    onPaint: {
        const ctx = getContext("2d");
        ctx.reset();

        const w = width;
        const h = height;
        if (w <= 0 || h <= 0) {
            return;
        }

        const y = v => h - (Math.min(100, Math.max(0, v)) / 100) * (h - lineWidth) - lineWidth / 2;
        const step = capacity > 1 ? w / (capacity - 1) : w;

        ctx.lineWidth = 1;
        ctx.strokeStyle = guideColor;
        ctx.globalAlpha = 0.3;
        for (const g of guides) {
            ctx.beginPath();
            ctx.moveTo(0, Math.round(y(g)) + 0.5);
            ctx.lineTo(w, Math.round(y(g)) + 0.5);
            ctx.stroke();
        }

        for (const s of series) {
            const values = s.values || [];
            if (values.length === 0) {
                continue;
            }
            const x0 = w - (values.length - 1) * step;

            ctx.beginPath();
            values.forEach((v, i) => {
                const x = x0 + i * step;
                if (i === 0) {
                    ctx.moveTo(x, y(v));
                } else {
                    ctx.lineTo(x, y(v));
                }
            });

            ctx.globalAlpha = 1;
            ctx.lineWidth = lineWidth;
            ctx.lineJoin = "round";
            ctx.strokeStyle = s.color;
            ctx.stroke();

            if (fill && values.length > 1) {
                ctx.lineTo(w, h);
                ctx.lineTo(x0, h);
                ctx.closePath();
                ctx.globalAlpha = 0.15;
                ctx.fillStyle = s.color;
                ctx.fill();
            }
        }
    }
}
