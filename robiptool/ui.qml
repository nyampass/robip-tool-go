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

   Component {
 id: highlight
	 Rectangle {
   width: 180; height: 40
				 color: "lightsteelblue"; radius: 5
											y: list.currentItem.y
											Behavior on y {
	   SpringAnimation {
	   spring: 3
		   damping: 0.2
		   }
	 }
   }
 }

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
		 onClicked: {
		 binding.onClicked()
		   }
	 }
   }
   }

   GroupBox {
   id: portBox
	   title: "Port"
	 Layout.fillWidth: true
	 Layout.minimumHeight: 80

	 ComboBox {
	 property var pseudoModel: []
	   model: pseudoModel
	   Component.onCompleted: {
	   var i;
	   for(i=0; i<binding.portsLength(); i++){
		 pseudoModel.push(binding.portAt(i))
		   }
	   model = pseudoModel
		 currentIndex = CurrentProfile
		 }
   onCurrentIndexChanged: binding.onSelectPort(currentIndex)
	   width: 200
	   }
   }
   TextArea {
   id: log
	   text: ""
	 Layout.minimumHeight: 30
	 Layout.fillWidth: true
	 }
 }
}
