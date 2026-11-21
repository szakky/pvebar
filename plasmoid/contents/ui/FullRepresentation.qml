import QtQuick
import QtQuick.Controls as QQC2
import QtQuick.Layouts
import org.kde.plasma.plasmoid
import org.kde.plasma.components as PlasmaComponents
import org.kde.plasma.extras as PlasmaExtras
import org.kde.kirigami as Kirigami

// Popup view: usage bars, a larger history graph and the guest list.
PlasmaExtras.Representation {
    id: full

    required property var main

    readonly property var status: main.status
    readonly property color cpuColor: Kirigami.Theme.highlightColor
    readonly property color memColor: Kirigami.Theme.positiveTextColor

    Layout.minimumWidth: Kirigami.Units.gridUnit * 18
    Layout.preferredWidth: Kirigami.Units.gridUnit * 22
    Layout.minimumHeight: Kirigami.Units.gridUnit * 16
    Layout.preferredHeight: Kirigami.Units.gridUnit * 24

    function formatBytes(bytes) {
        const gib = bytes / (1024 * 1024 * 1024);
        return gib >= 10 ? gib.toFixed(0) + " GiB" : gib.toFixed(1) + " GiB";
    }

    header: PlasmaExtras.PlasmoidHeading {
        RowLayout {
            anchors.fill: parent

            Kirigami.Heading {
                text: "Proxmox VE"
                level: 3
                Layout.fillWidth: true
            }

            PlasmaComponents.Label {
                visible: full.status !== null
                text: i18n("Updated %1", full.main.lastUpdate.toLocaleTimeString(Qt.locale(), Locale.ShortFormat))
                opacity: 0.7
                font: Kirigami.Theme.smallFont
            }

            PlasmaComponents.ToolButton {
                icon.name: "view-refresh"
                onClicked: full.main.refresh()
                QQC2.ToolTip.text: i18n("Refresh now")
                QQC2.ToolTip.visible: hovered
            }
        }
    }

    ColumnLayout {
        anchors.fill: parent
        spacing: Kirigami.Units.largeSpacing

        Kirigami.InlineMessage {
            Layout.fillWidth: true
            type: Kirigami.MessageType.Error
            text: full.main.errorText
            visible: full.main.hasError
        }

        PlasmaExtras.PlaceholderMessage {
            Layout.fillWidth: true
            Layout.fillHeight: true
            visible: full.status === null && !full.main.hasError
            iconName: "network-server"
            text: i18n("Waiting for pvebar…")
        }

        // CPU and RAM bars
        GridLayout {
            visible: full.status !== null
            Layout.fillWidth: true
            columns: 3
            columnSpacing: Kirigami.Units.smallSpacing

            Rectangle {
                width: Kirigami.Units.smallSpacing * 2
                height: width
                radius: width / 2
                color: full.cpuColor
            }
            PlasmaComponents.Label {
                text: i18n("CPU %1%", full.status ? full.status.cpu.toFixed(1) : "–")
                color: full.status ? full.main.levelColor(full.status.cpu, Kirigami.Theme.textColor, Kirigami.Theme)
                                   : Kirigami.Theme.textColor
                Layout.minimumWidth: Kirigami.Units.gridUnit * 5
            }
            PlasmaComponents.ProgressBar {
                Layout.fillWidth: true
                from: 0
                to: 100
                value: full.status ? full.status.cpu : 0
            }

            Rectangle {
                width: Kirigami.Units.smallSpacing * 2
                height: width
                radius: width / 2
                color: full.memColor
            }
            PlasmaComponents.Label {
                text: i18n("RAM %1%", full.status ? full.status.mem.toFixed(1) : "–")
                color: full.status ? full.main.levelColor(full.status.mem, Kirigami.Theme.textColor, Kirigami.Theme)
                                   : Kirigami.Theme.textColor
            }
            PlasmaComponents.ProgressBar {
                Layout.fillWidth: true
                from: 0
                to: 100
                value: full.status ? full.status.mem : 0
            }

            Item { width: 1; height: 1 }
            Item { width: 1; height: 1 }
            PlasmaComponents.Label {
                text: full.status ? i18n("%1 / %2", full.formatBytes(full.status.mem_used),
                                         full.formatBytes(full.status.mem_total)) : ""
                font: Kirigami.Theme.smallFont
                opacity: 0.7
                horizontalAlignment: Text.AlignRight
                Layout.fillWidth: true
            }
        }

        // History graph
        ColumnLayout {
            visible: full.status !== null
            Layout.fillWidth: true
            spacing: 0

            Rectangle {
                Layout.fillWidth: true
                Layout.preferredHeight: Kirigami.Units.gridUnit * 5
                color: "transparent"
                border.color: Qt.rgba(Kirigami.Theme.textColor.r, Kirigami.Theme.textColor.g,
                                      Kirigami.Theme.textColor.b, 0.2)
                radius: Kirigami.Units.cornerRadius

                Sparkline {
                    anchors.fill: parent
                    anchors.margins: 1
                    capacity: full.main.historySize
                    guides: [50, full.main.warn]
                    guideColor: Kirigami.Theme.textColor
                    series: [
                        { values: full.main.memHistory, color: full.memColor },
                        { values: full.main.cpuHistory, color: full.cpuColor }
                    ]
                }
            }

            PlasmaComponents.Label {
                Layout.fillWidth: true
                text: i18n("Last %1 min (%2 samples)",
                           Math.round(full.main.historySize * Plasmoid.configuration.interval / 60),
                           full.main.cpuHistory.length)
                font: Kirigami.Theme.smallFont
                opacity: 0.7
                horizontalAlignment: Text.AlignRight
            }
        }

        Kirigami.Heading {
            visible: full.status !== null
            text: i18n("Guests %1/%2", full.status ? full.status.running : 0, full.status ? full.status.total : 0)
            level: 5
        }

        // Guest list
        PlasmaComponents.ScrollView {
            visible: full.status !== null
            Layout.fillWidth: true
            Layout.fillHeight: true

            ListView {
                model: full.status ? full.status.guests : []
                spacing: Kirigami.Units.smallSpacing
                clip: true

                delegate: RowLayout {
                    id: guestRow

                    required property var modelData
                    readonly property bool running: modelData.status === "running"

                    width: ListView.view.width
                    spacing: Kirigami.Units.smallSpacing

                    Rectangle {
                        width: Kirigami.Units.smallSpacing * 2
                        height: width
                        radius: width / 2
                        color: guestRow.running ? Kirigami.Theme.positiveTextColor : Kirigami.Theme.disabledTextColor
                        Layout.alignment: Qt.AlignVCenter
                    }
                    PlasmaComponents.Label {
                        text: guestRow.modelData.vmid + " " + guestRow.modelData.name
                        elide: Text.ElideRight
                        Layout.fillWidth: true
                    }
                    PlasmaComponents.Label {
                        text: guestRow.modelData.type === "lxc" ? "CT" : "VM"
                        opacity: 0.6
                        font: Kirigami.Theme.smallFont
                    }
                    PlasmaComponents.Label {
                        text: guestRow.running
                            ? i18n("CPU %1%  %2", guestRow.modelData.cpu.toFixed(0),
                                   full.formatBytes(guestRow.modelData.mem))
                            : guestRow.modelData.status
                        opacity: guestRow.running ? 1 : 0.6
                        horizontalAlignment: Text.AlignRight
                        Layout.minimumWidth: Kirigami.Units.gridUnit * 6
                    }
                }
            }
        }
    }
}
