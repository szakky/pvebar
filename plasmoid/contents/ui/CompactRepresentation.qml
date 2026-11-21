import QtQuick
import QtQuick.Layouts
import org.kde.plasma.plasmoid
import org.kde.plasma.core as PlasmaCore
import org.kde.plasma.components as PlasmaComponents
import org.kde.kirigami as Kirigami

// Panel view: a small CPU/RAM graph and the running guest count.
MouseArea {
    id: compact

    required property var main

    readonly property bool vertical: Plasmoid.formFactor === PlasmaCore.Types.Vertical
    readonly property color cpuColor: Kirigami.Theme.highlightColor
    readonly property color memColor: Kirigami.Theme.positiveTextColor

    Layout.minimumWidth: vertical ? -1 : content.implicitWidth
    Layout.preferredWidth: Layout.minimumWidth
    Layout.minimumHeight: vertical ? content.implicitHeight : -1

    hoverEnabled: true
    onClicked: compact.main.expanded = !compact.main.expanded

    RowLayout {
        id: content
        anchors.fill: parent
        spacing: Kirigami.Units.smallSpacing

        Kirigami.Icon {
            visible: compact.main.hasError
            source: "dialog-warning"
            Layout.preferredWidth: Kirigami.Units.iconSizes.small
            Layout.preferredHeight: Kirigami.Units.iconSizes.small
            Layout.alignment: Qt.AlignVCenter
        }

        Sparkline {
            visible: !compact.main.hasError
            Layout.preferredWidth: Kirigami.Units.gridUnit * 3
            Layout.fillHeight: true
            Layout.topMargin: Kirigami.Units.smallSpacing
            Layout.bottomMargin: Kirigami.Units.smallSpacing
            capacity: Math.min(compact.main.historySize, 40)
            lineWidth: 1
            series: [
                { values: compact.main.memHistory.slice(-40), color: compact.memColor },
                { values: compact.main.cpuHistory.slice(-40), color: compact.cpuColor }
            ]
        }

        PlasmaComponents.Label {
            text: compact.main.hasError ? "PVE !"
                : compact.main.status ? compact.main.status.running + "/" + compact.main.status.total
                : "PVE"
            color: compact.main.status && !compact.main.hasError
                ? compact.main.levelColor(Math.max(compact.main.status.cpu, compact.main.status.mem),
                                          Kirigami.Theme.textColor, Kirigami.Theme)
                : Kirigami.Theme.textColor
            Layout.alignment: Qt.AlignVCenter
        }
    }
}
