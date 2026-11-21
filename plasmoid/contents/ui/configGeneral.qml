import QtQuick
import QtQuick.Controls as QQC2
import QtQuick.Layouts
import org.kde.kirigami as Kirigami
import org.kde.kcmutils as KCM

KCM.SimpleKCM {
    property alias cfg_command: commandField.text
    property alias cfg_interval: intervalSpin.value
    property alias cfg_historySize: historySpin.value
    property alias cfg_warn: warnSpin.value
    property alias cfg_crit: critSpin.value

    Kirigami.FormLayout {
        QQC2.TextField {
            id: commandField
            Kirigami.FormData.label: i18n("pvebar command:")
            placeholderText: "$HOME/go/bin/pvebar"
        }

        QQC2.Label {
            text: i18n("Use the full path; panels do not read ~/.bashrc. \"--format json\" is added automatically.")
            wrapMode: Text.WordWrap
            Layout.maximumWidth: Kirigami.Units.gridUnit * 20
            font: Kirigami.Theme.smallFont
            opacity: 0.7
        }

        QQC2.SpinBox {
            id: intervalSpin
            Kirigami.FormData.label: i18n("Update every (seconds):")
            from: 5
            to: 3600
        }

        QQC2.SpinBox {
            id: historySpin
            Kirigami.FormData.label: i18n("Samples in graph:")
            from: 10
            to: 2000
        }

        QQC2.SpinBox {
            id: warnSpin
            Kirigami.FormData.label: i18n("Warning at (%):")
            from: 0
            to: 100
        }

        QQC2.SpinBox {
            id: critSpin
            Kirigami.FormData.label: i18n("Critical at (%):")
            from: 0
            to: 100
        }
    }
}
