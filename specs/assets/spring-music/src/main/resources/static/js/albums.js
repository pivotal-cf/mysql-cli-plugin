angular.module('albums', ['ngResource', 'ui.bootstrap']).
    factory('Albums', function ($resource) {
        return $resource('albums');
    }).
    factory('Album', function ($resource) {
        return $resource('albums/:id', { id: '@id' });
    }).
    factory("EditorStatus", function () {
        var editorEnabled = {};

        var enable = function (id, fieldName) {
            editorEnabled = { 'id': id, 'fieldName': fieldName };
        };

        var disable = function () {
            editorEnabled = {};
        };

        var isEnabled = function (id, fieldName) {
            return (editorEnabled['id'] == id && editorEnabled['fieldName'] == fieldName);
        };

        return {
            isEnabled: isEnabled,
            enable: enable,
            disable: disable
        }
    });

function AlbumsController($scope, Albums, Album, Status) {
    function list() {
        $scope.albums = Albums.query();
    }

    function clone(obj) {
        return JSON.parse(JSON.stringify(obj));
    }

    function saveAlbum(album) {
        Albums.save(album,
            function () {
                Status.success("Album saved");
                list();
            },
            function (result) {
                Status.error("Error saving album: " + result.status);
            }
        );
    }

    $scope.addAlbum = function () {
        $scope.showModal = true;
        $scope.modalAlbum = {};
        $scope.modalAction = 'add';
    };

    $scope.updateAlbum = function (album) {
        $scope.showModal = true;
        $scope.modalAlbum = clone(album);
        $scope.modalAction = 'update';
    };

    $scope.deleteAlbum = function (album) {
        Album.delete({ id: album.id },
            function () {
                Status.success("Album deleted");
                list();
            },
            function (result) {
                Status.error("Error deleting album: " + result.status);
            }
        );
    };

    $scope.setAlbumsView = function (viewName) {
        console.log("Setting albums view to:", viewName);
        $scope.albumsView = "templates/" + viewName + ".html";
        console.log("Albums view path:", $scope.albumsView);
    };

    $scope.init = function () {
        list();
        $scope.setAlbumsView("grid");
        $scope.sortField = "title";
        $scope.sortDescending = false;
        $scope.showModal = false;
    };

    $scope.saveModalAlbum = function () {
        if ($scope.albumForm.$valid) {
            saveAlbum($scope.modalAlbum);
            $scope.closeModal();
        }
    };

    $scope.closeModal = function () {
        $scope.showModal = false;
        $scope.modalAlbum = {};
    };
}

function AlbumModalController($scope, $modalInstance, album, action) {
    $scope.albumAction = action;
    $scope.yearPattern = /^[1-2]\d{3}$/;
    $scope.album = album;

    $scope.ok = function () {
        $modalInstance.close($scope.album);
    };

    $scope.cancel = function () {
        $modalInstance.dismiss('cancel');
    };
};

function AlbumEditorController($scope, Albums, Status, EditorStatus) {
    $scope.enableEditor = function (album, fieldName) {
        $scope.newFieldValue = album[fieldName];
        EditorStatus.enable(album.id, fieldName);
    };

    $scope.disableEditor = function () {
        EditorStatus.disable();
    };

    $scope.isEditorEnabled = function (album, fieldName) {
        return EditorStatus.isEnabled(album.id, fieldName);
    };

    $scope.save = function (album, fieldName) {
        if ($scope.newFieldValue === "") {
            return false;
        }

        album[fieldName] = $scope.newFieldValue;

        Albums.save({}, album,
            function () {
                Status.success("Album saved");
                list();
            },
            function (result) {
                Status.error("Error saving album: " + result.status);
            }
        );

        $scope.disableEditor();
    };

    $scope.disableEditor();
}

angular.module('albums').
    directive('inPlaceEdit', function () {
        return {
            restrict: 'E',
            transclude: true,
            replace: true,

            scope: {
                ipeFieldName: '@fieldName',
                ipeInputType: '@inputType',
                ipeInputClass: '@inputClass',
                ipePattern: '@pattern',
                ipeModel: '=model'
            },

            template:
                '<div>' +
                '<span ng-hide="isEditorEnabled(ipeModel, ipeFieldName)" ng-click="enableEditor(ipeModel, ipeFieldName)">' +
                '<span ng-transclude></span>' +
                '</span>' +
                '<span ng-show="isEditorEnabled(ipeModel, ipeFieldName)">' +
                '<div class="input-group input-group-sm">' +
                '<input type="{{ipeInputType}}" name="{{ipeFieldName}}" class="form-control" ' +
                'ng-required ng-pattern="{{ipePattern}}" ng-model="newFieldValue" ' +
                'ui-keyup="{enter: \'save(ipeModel, ipeFieldName)\', esc: \'disableEditor()\'}"/>' +
                '<button ng-click="save(ipeModel, ipeFieldName)" type="button" class="btn btn-success btn-sm"><span class="glyphicon glyphicon-ok"></span></button>' +
                '<button ng-click="disableEditor()" type="button" class="btn btn-secondary btn-sm"><span class="glyphicon glyphicon-remove"></span></button>' +
                '</div>' +
                '</span>' +
                '</div>',

            controller: 'AlbumEditorController'
        };
    }).
    directive('bootstrapDropdown', function () {
        return {
            restrict: 'A',
            link: function (scope, element, attrs) {
                // Handle dropdown clicks manually for AngularJS compatibility
                element.on('click', '.dropdown-item', function (e) {
                    e.preventDefault();
                    e.stopPropagation();

                    // Close the dropdown
                    var dropdown = element.find('.dropdown-menu');
                    dropdown.removeClass('show');
                    element.find('.dropdown-toggle').attr('aria-expanded', 'false');

                    // Execute the ng-click function
                    var clickHandler = angular.element(this).attr('ng-click');
                    if (clickHandler) {
                        scope.$eval(clickHandler);
                        scope.$apply();
                    }
                });
            }
        };
    });
