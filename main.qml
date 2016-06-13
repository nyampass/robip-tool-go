import QtQuick 2.2
import QtQuick.Controls 1.1
import QtQuick.Layouts 1.0

ApplicationWindow {
    visible: true
    title: "Robip tool 1.2"
    property int margin: 11
    width: mainLayout.implicitWidth + 2 * margin
    height: mainLayout.implicitHeight + 2 * margin
    minimumWidth: mainLayout.Layout.minimumWidth + 2 * margin
    minimumHeight: mainLayout.Layout.minimumHeight + 2 * margin

    ColumnLayout {
        id: mainLayout
        anchors.fill: parent
        anchors.margins: margin
        GroupBox {
            id: rowBox
            title: "Robip ID"
            Layout.fillWidth: true

            RowLayout {
                id: rowLayout
                anchors.fill: parent
                TextField {
                    placeholderText: "Robip ID"
                    Layout.fillWidth: true
                }
                Button {
                    text: "書き込み"
                }
            }
        }

        GroupBox {
            id: gridBox
            Layout.fillWidth: true

            GridLayout {
                id: gridLayout
                rows: 3
                flow: GridLayout.TopToBottom
                anchors.fill: parent

                Label { text: "Line 1" }
                Label { text: "Line 2" }
                Label { text: "Line 3" }

                TextField { }
                TextField { }
                TextField { }

                TextArea {
                    text: "This widget spans over three rows in the GridLayout.\n"
                        + "All items in the GridLayout are implicitly positioned from top to bottom."
                    Layout.rowSpan: 3
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                }
            }
        }
        TextArea {
            id: log
            text: ""
            Layout.minimumHeight: 30
            Layout.fillHeight: true
            Layout.fillWidth: true
        }
    }
}