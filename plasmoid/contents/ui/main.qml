import QtQuick
import org.kde.plasma.plasmoid
import org.kde.plasma.core as PlasmaCore
import org.kde.plasma.plasma5support as P5Support

// pvebar widget: runs "pvebar --format json" on a timer and keeps a short
// history in memory for the graphs. The history is lost when Plasma restarts.
PlasmoidItem {
    id: root

    // Last successful result from pvebar (see "JSON output" in README.md).
    property var status: null
    property string errorText: ""
    property date lastUpdate

    // Percent values, oldest first.
    property var cpuHistory: []
    property var memHistory: []

    readonly property int historySize: Plasmoid.configuration.historySize
    readonly property int warn: Plasmoid.configuration.warn
    readonly property int crit: Plasmoid.configuration.crit
    readonly property bool hasError: errorText !== ""

    Plasmoid.icon: "network-server"

    toolTipMainText: "Proxmox VE"
    toolTipSubText: {
        if (hasError) {
            return i18n("Error: %1", errorText);
        }
        if (!status) {
            return i18n("Waiting for pvebar…");
        }
        return i18n("CPU %1%  RAM %2%  Guests %3/%4",
                    status.cpu.toFixed(1), status.mem.toFixed(1),
                    status.running, status.total);
    }

    compactRepresentation: CompactRepresentation {
        main: root
    }

    fullRepresentation: FullRepresentation {
        main: root
    }

    // Color for a usage percentage, following the warn/crit thresholds.
    function levelColor(value, normalColor, theme) {
        if (value >= crit) {
            return theme.negativeTextColor;
        }
        if (value >= warn) {
            return theme.neutralTextColor;
        }
        return normalColor;
    }

    function pushSample(list, value) {
        // Assign a new array so bindings and the graphs update.
        const next = list.slice(Math.max(0, list.length - historySize + 1));
        next.push(value);
        return next;
    }

    function handleOutput(stdout, stderr, exitCode) {
        let result;
        try {
            result = JSON.parse(stdout);
        } catch (e) {
            // Not JSON: pvebar is missing, too old, or crashed.
            const detail = (stderr || stdout || "").trim();
            errorText = detail !== "" ? detail : i18n("pvebar exited with code %1", exitCode);
            return;
        }

        if (!result.ok) {
            errorText = result.error || i18n("unknown error");
            return;
        }

        errorText = "";
        status = result;
        lastUpdate = new Date();
        cpuHistory = pushSample(cpuHistory, result.cpu);
        memHistory = pushSample(memHistory, result.mem);
    }

    function refresh() {
        executable.connectSource(Plasmoid.configuration.command + " --format json");
    }

    P5Support.DataSource {
        id: executable
        engine: "executable"
        connectedSources: []

        onNewData: (sourceName, data) => {
            // Disconnect so the same command can run again on the next tick.
            disconnectSource(sourceName);
            root.handleOutput(data["stdout"], data["stderr"], data["exit code"]);
        }
    }

    Timer {
        interval: Math.max(5, Plasmoid.configuration.interval) * 1000
        running: true
        repeat: true
        triggeredOnStart: true
        onTriggered: root.refresh()
    }

    Plasmoid.contextualActions: [
        PlasmaCore.Action {
            text: i18n("Refresh now")
            icon.name: "view-refresh"
            onTriggered: root.refresh()
        }
    ]
}
