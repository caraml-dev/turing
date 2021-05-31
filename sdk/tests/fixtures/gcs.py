import json
import os
import pytest
import google.auth.environment_vars


@pytest.fixture
def mock_gcs(responses, bucket_name):
    os.environ[google.auth.environment_vars.PROJECT] = "test-project"

    responses.add(
        method="POST",
        url="/token",
        body=json.dumps({
            "access_token": "ya29.ImCpB6BS2mdOMseaUjhVlHqNfAOz168XjuDrK7Sd33glPd7XvtMLIngi1-V52ReytFSUluE-iBV88OlDkjtraggB_qc-LN2JlGtQ3sHZq_MuTxrU0-oK_kpq-1wsvniFFGQ",
            "expires_in": 3600,
            "scope": "openid https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email",
            "token_type": "Bearer",
            "id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjhhNjNmZTcxZTUzMDY3NTI0Y2JiYzZhM2E1ODQ2M2IzODY0YzA3ODciLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiI3NjQwODYwNTE4NTAtNnFyNHA2Z3BpNmhuNTA2cHQ4ZWp1cTgzZGkzNDFodXIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJhdWQiOiI3NjQwODYwNTE4NTAtNnFyNHA2Z3BpNmhuNTA2cHQ4ZWp1cTgzZGkzNDFodXIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJzdWIiOiIxMDM5ODg0MzM2OTY3NzI1NDkzNjAiLCJoZCI6ImdvLWplay5jb20iLCJlbWFpbCI6InByYWRpdGh5YS5wdXJhQGdvLWplay5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiYXRfaGFzaCI6ImdrbXIxY0dPTzNsT0dZUDhtYjNJRnciLCJpYXQiOjE1NzE5Njg3NDUsImV4cCI6MTU3MTk3MjM0NX0.FIY5xvySNVxt1cbw-QXdDfiwollxcqupz1YDJuP14obKRyDwFG9ZcC_j-mTDZF5_dzpYeNMMK-LPTq9QIaM-blSKm2Eh9LeMvQGUk_S-9y_r2jKCmBlrEeHM8DXk3xyKf65LEoBA8cwMPdgb2s8AMIxxN9JJ09fjou20yLDI84Q4BFMriMIBBYLFgBW0wcg2PQ1hy5hrV1PdZj-ZNKNWmouh0lOjLLYmVFZPCzD9ENWo1N52ZLaLODdI2gDcpbyTUbeAh81sacdtJd0pLf-FuBLdfuktvP4MVvdmIhXv98Zb0dFBzRtmiqlQusSjoG5VEaBc6o2gkM5rHR0ozby0Fg"
        }),
        status=200,
        content_type='application/json'
    )

    responses.add(
        method="POST",
        url=f"/upload/storage/v1/b/{bucket_name}/o?uploadType=multipart",
        match_querystring=True,
        body="{}",
        status=200,
        content_type='multipart/form-data'
    )
